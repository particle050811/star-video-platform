package chat

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	chat "video-platform/biz/model/chat"
	"video-platform/biz/service"
	"video-platform/pkg/middleware"
	"video-platform/pkg/parser"
	"video-platform/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/websocket"
)

const (
	chatWSWriteWait         = 10 * time.Second
	chatWSPongWait          = 60 * time.Second
	chatWSPingPeriod        = 45 * time.Second
	chatWSMaxMessageSize    = 4096
	chatWSSendBufferSize    = 256
	chatWSSubscribeRetry    = 3 * time.Second
	chatWSPublishRetries    = 3
	chatWSPublishRetryDelay = 100 * time.Millisecond
)

const (
	chatWSEventConnected      = "connected"
	chatWSEventError          = "error"
	chatWSEventPing           = "ping"
	chatWSEventPong           = "pong"
	chatWSEventSendMessage    = "send_message"
	chatWSEventMessageCreated = "message_created"
	chatWSEventMarkRead       = "mark_read"
	chatWSEventReadMarked     = "read_marked"
)

type chatWSEnvelope struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data,omitempty"`
}

type chatWSSendMessageRequest struct {
	RoomID      string `json:"room_id"`
	Content     string `json:"content"`
	ClientMsgID string `json:"client_msg_id"`
}

type chatWSMarkReadRequest struct {
	RoomID            string `json:"room_id"`
	LastReadMessageID string `json:"last_read_message_id"`
}

type chatWSErrorPayload struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type chatWSReadMarkedPayload struct {
	RoomID            string `json:"room_id"`
	LastReadMessageID string `json:"last_read_message_id"`
}

type chatWSMessageEvent struct {
	Origin        string `json:"origin"`
	MemberUserIDs []uint `json:"member_user_ids"`
	Message       any    `json:"message"`
}

type chatWSClient struct {
	userID uint
	conn   *websocket.Conn
	send   chan any
	mu     sync.Mutex
	closed bool
}

type chatWSHub struct {
	mu      sync.RWMutex
	clients map[uint]map[*chatWSClient]struct{}
}

var (
	chatWSUpgrader = websocket.HertzUpgrader{
		CheckOrigin: func(ctx *app.RequestContext) bool {
			return true
		},
	}
	defaultChatWSHub          = newChatWSHub()
	chatWSSubscriberOnce      sync.Once
	chatWSSubscribeEvents     = service.Chat.SubscribeMessageEvents
	chatWSWaitSubscribeRetry  = sleepChatWSSubscribeRetry
	chatWSPublishMessageEvent = service.Chat.PublishMessageEvent
	chatWSWaitPublishRetry    = sleepChatWSPublishRetry
)

func newChatWSHub() *chatWSHub {
	return &chatWSHub{
		clients: make(map[uint]map[*chatWSClient]struct{}),
	}
}

func (h *chatWSHub) register(client *chatWSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.userID] == nil {
		h.clients[client.userID] = make(map[*chatWSClient]struct{})
	}
	h.clients[client.userID][client] = struct{}{}
}

func (h *chatWSHub) unregister(client *chatWSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userClients := h.clients[client.userID]
	if userClients == nil {
		return
	}
	if _, ok := userClients[client]; !ok {
		return
	}
	delete(userClients, client)
	client.closeSend()
	if len(userClients) == 0 {
		delete(h.clients, client.userID)
	}
}

func (h *chatWSHub) sendToUser(userID uint, payload any) {
	h.mu.RLock()
	slowClients := make([]*chatWSClient, 0)
	for client := range h.clients[userID] {
		if !client.trySend(payload) {
			slowClients = append(slowClients, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range slowClients {
		h.unregister(client)
	}
}

func (h *chatWSHub) sendToClient(client *chatWSClient, payload any) bool {
	h.mu.RLock()
	userClients := h.clients[client.userID]
	if _, ok := userClients[client]; !ok {
		h.mu.RUnlock()
		return false
	}

	sent := client.trySend(payload)
	h.mu.RUnlock()

	if !sent {
		h.unregister(client)
	}
	return sent
}

func (c *chatWSClient) trySend(payload any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return false
	}
	select {
	case c.send <- payload:
		return true
	default:
		return false
	}
}

func (c *chatWSClient) closeSend() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}
	c.closed = true
	close(c.send)
}

func (h *chatWSHub) broadcastToUsers(userIDs []uint, payload any) {
	for _, userID := range userIDs {
		h.sendToUser(userID, payload)
	}
}

func StartChatWebSocketSubscriber(ctx context.Context) {
	chatWSSubscriberOnce.Do(func() {
		go subscribeChatMessageEvents(ctx)
	})
}

// ChatWebSocket upgrades an authenticated chat request to a websocket connection.
func ChatWebSocket(ctx context.Context, c *app.RequestContext) {
	userIDValue, _ := c.Get(middleware.ContextUserID)
	userID := userIDValue.(uint)

	if err := chatWSUpgrader.Upgrade(c, func(conn *websocket.Conn) {
		client := &chatWSClient{
			userID: userID,
			conn:   conn,
			send:   make(chan any, chatWSSendBufferSize),
		}
		defaultChatWSHub.register(client)
		defer defaultChatWSHub.unregister(client)

		go client.writePump()
		defaultChatWSHub.sendToClient(client, chatWSResponse(chatWSEventConnected, map[string]string{"user_id": stringFromUint(userID)}))
		client.readPump(ctx)
	}); err != nil {
		log.Printf("[聊天模块][WebSocket] 升级连接失败 user_id=%d: %v", userID, err)
	}
}

func (c *chatWSClient) readPump(ctx context.Context) {
	defer func() {
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(chatWSMaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(chatWSPongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(chatWSPongWait))
	})

	for {
		var envelope chatWSEnvelope
		if err := c.conn.ReadJSON(&envelope); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[聊天模块][WebSocket] 读取消息失败 user_id=%d: %v", c.userID, err)
			}
			return
		}
		c.handleEnvelope(ctx, envelope)
	}
}

func (c *chatWSClient) writePump() {
	ticker := time.NewTicker(chatWSPingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case payload, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(chatWSWriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(chatWSWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *chatWSClient) handleEnvelope(ctx context.Context, envelope chatWSEnvelope) {
	switch envelope.Event {
	case chatWSEventPing:
		defaultChatWSHub.sendToClient(c, chatWSResponse(chatWSEventPong, nil))
	case chatWSEventSendMessage:
		c.handleSendMessage(ctx, envelope.Data)
	case chatWSEventMarkRead:
		c.handleMarkRead(ctx, envelope.Data)
	default:
		c.sendError(consts.StatusBadRequest, "不支持的 websocket 事件")
	}
}

func (c *chatWSClient) handleSendMessage(ctx context.Context, raw json.RawMessage) {
	var req chatWSSendMessageRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		c.sendError(consts.StatusBadRequest, "消息格式错误")
		return
	}

	roomID, err := parser.ChatRoomID(req.RoomID)
	if err != nil {
		c.sendError(consts.StatusBadRequest, err.Error())
		return
	}

	memberIDs, err := service.Chat.ListMemberUserIDs(ctx, c.userID, roomID)
	if err != nil {
		c.writeServiceError(roomID, err)
		return
	}

	message, err := service.Chat.CreateMessage(ctx, c.userID, roomID, req.Content, req.ClientMsgID)
	if err != nil {
		c.writeServiceError(roomID, err)
		return
	}
	payload := chatWSResponse(chatWSEventMessageCreated, message)
	defaultChatWSHub.broadcastToUsers(memberIDs, payload)
	if err := publishChatMessageEventWithRetry(ctx, memberIDs, message); err != nil {
		log.Printf("[聊天模块][WebSocket] Redis 发布消息事件失败 room_id=%d message_id=%s: %v", roomID, message.Id, err)
	}
}

func (c *chatWSClient) handleMarkRead(ctx context.Context, raw json.RawMessage) {
	var req chatWSMarkReadRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		c.sendError(consts.StatusBadRequest, "消息格式错误")
		return
	}

	roomID, err := parser.ChatRoomID(req.RoomID)
	if err != nil {
		c.sendError(consts.StatusBadRequest, err.Error())
		return
	}
	messageID, err := parser.ChatMessageID(req.LastReadMessageID)
	if err != nil {
		c.sendError(consts.StatusBadRequest, err.Error())
		return
	}

	if err := service.Chat.MarkRoomRead(ctx, c.userID, roomID, messageID); err != nil {
		c.writeServiceError(roomID, err)
		return
	}

	defaultChatWSHub.sendToClient(c, chatWSResponse(chatWSEventReadMarked, chatWSReadMarkedPayload(req)))
}

func (c *chatWSClient) writeServiceError(roomID uint, err error) {
	switch {
	case errors.Is(err, service.ErrChatRoomNotFound):
		c.sendError(response.CodeChatRoomNotFound, "聊天房间不存在")
	case errors.Is(err, service.ErrChatRoomMemberNotFound):
		c.sendError(response.CodeChatRoomMemberNotFound, "不是聊天房间成员")
	case errors.Is(err, service.ErrChatMessageNotFound):
		c.sendError(response.CodeChatMessageNotFound, "聊天消息不存在")
	case errors.Is(err, service.ErrCommentEmpty):
		c.sendError(consts.StatusBadRequest, "消息内容不能为空")
	default:
		log.Printf("[聊天模块][WebSocket] 操作失败 user_id=%d room_id=%d: %v", c.userID, roomID, err)
		c.sendError(consts.StatusInternalServerError, "服务器内部错误")
	}
}

func (c *chatWSClient) sendError(code int32, message string) {
	defaultChatWSHub.sendToClient(c, chatWSResponse(chatWSEventError, chatWSErrorPayload{
		Code:    code,
		Message: message,
	}))
}

func chatWSResponse(event string, data any) map[string]any {
	return map[string]any{
		"event": event,
		"data":  data,
	}
}

func stringFromUint(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

func subscribeChatMessageEvents(ctx context.Context) {
	for {
		if err := subscribeChatMessageEventsOnce(ctx); err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[聊天模块][WebSocket] Redis 消息订阅中断，准备重试: %v", err)
		}
		if !chatWSWaitSubscribeRetry(ctx) {
			return
		}
	}
}

func subscribeChatMessageEventsOnce(ctx context.Context) error {
	events, closeSubscription, err := chatWSSubscribeEvents(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if closeSubscription != nil {
			_ = closeSubscription()
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case raw, ok := <-events:
			if !ok {
				return errors.New("redis chat message subscription closed")
			}
			handleChatMessageEvent(ctx, raw)
		}
	}
}

func sleepChatWSSubscribeRetry(ctx context.Context) bool {
	timer := time.NewTimer(chatWSSubscribeRetry)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func publishChatMessageEventWithRetry(ctx context.Context, memberIDs []uint, message *chat.ChatMessage) error {
	var lastErr error
	for attempt := 0; attempt < chatWSPublishRetries; attempt++ {
		if err := chatWSPublishMessageEvent(ctx, memberIDs, message); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt < chatWSPublishRetries-1 && !chatWSWaitPublishRetry(ctx) {
			return ctx.Err()
		}
	}
	return lastErr
}

func sleepChatWSPublishRetry(ctx context.Context) bool {
	timer := time.NewTimer(chatWSPublishRetryDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func handleChatMessageEvent(ctx context.Context, raw string) {
	if err := ctx.Err(); err != nil {
		return
	}

	var event chatWSMessageEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		log.Printf("[聊天模块][WebSocket] Redis 消息事件解析失败: %v", err)
		return
	}
	if event.Origin == service.Chat.MessageEventOrigin() {
		return
	}
	defaultChatWSHub.broadcastToUsers(event.MemberUserIDs, chatWSResponse(chatWSEventMessageCreated, event.Message))
}
