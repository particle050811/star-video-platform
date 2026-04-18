package chat

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	modelchat "video-platform/biz/model/chat"
	"video-platform/biz/service"
)

func TestChatWSHubRegisterSendAndUnregister(t *testing.T) {
	hub := newChatWSHub()
	client := newTestChatWSClient(1, 1)

	hub.register(client)
	if !hub.sendToClient(client, chatWSResponse(chatWSEventPong, nil)) {
		t.Fatal("expected sendToClient to succeed for registered client")
	}

	payload := receiveChatWSPayload(t, client.send)
	if payload["event"] != chatWSEventPong {
		t.Fatalf("unexpected event: %v", payload["event"])
	}

	hub.unregister(client)
	if hub.sendToClient(client, chatWSResponse(chatWSEventPong, nil)) {
		t.Fatal("expected sendToClient to fail after unregister")
	}
	assertChannelClosed(t, client.send)
}

func TestChatWSHubBroadcastToMultipleConnections(t *testing.T) {
	hub := newChatWSHub()
	first := newTestChatWSClient(1, 1)
	second := newTestChatWSClient(1, 1)
	otherUser := newTestChatWSClient(2, 1)

	hub.register(first)
	hub.register(second)
	hub.register(otherUser)

	hub.sendToUser(1, chatWSResponse(chatWSEventMessageCreated, map[string]string{"id": "101"}))

	for _, client := range []*chatWSClient{first, second} {
		payload := receiveChatWSPayload(t, client.send)
		if payload["event"] != chatWSEventMessageCreated {
			t.Fatalf("unexpected event: %v", payload["event"])
		}
	}

	select {
	case payload := <-otherUser.send:
		t.Fatalf("other user should not receive broadcast: %+v", payload)
	default:
	}
}

func TestChatWSHubUnregisterIsIdempotent(t *testing.T) {
	hub := newChatWSHub()
	client := newTestChatWSClient(1, 1)

	hub.register(client)
	hub.unregister(client)

	assertNotPanics(t, func() {
		hub.unregister(client)
	})
	assertChannelClosed(t, client.send)
}

func TestChatWSHubSlowClientIsUnregistered(t *testing.T) {
	hub := newChatWSHub()
	client := newTestChatWSClient(1, 1)

	hub.register(client)
	if !hub.sendToClient(client, chatWSResponse(chatWSEventPong, nil)) {
		t.Fatal("expected first send to fill client buffer")
	}
	if hub.sendToClient(client, chatWSResponse(chatWSEventPong, nil)) {
		t.Fatal("expected second send to unregister full client")
	}

	assertChannelDrainsAndCloses(t, client.send)
	if hub.sendToClient(client, chatWSResponse(chatWSEventPong, nil)) {
		t.Fatal("expected send to fail for unregistered slow client")
	}
}

func TestChatWSHubConcurrentRegisterSendAndUnregister(t *testing.T) {
	hub := newChatWSHub()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		userID := uint(i%10 + 1)
		wg.Add(1)
		go func(userID uint) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				client := newTestChatWSClient(userID, 16)
				hub.register(client)
				_ = hub.sendToClient(client, chatWSResponse(chatWSEventPong, nil))
				hub.sendToUser(client.userID, chatWSResponse(chatWSEventPong, nil))
				hub.unregister(client)
			}
		}(userID)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("concurrent hub operations timed out")
	}
}

func TestChatWSHubBroadcastsRedisMessageEvent(t *testing.T) {
	hub := newChatWSHub()
	originalHub := defaultChatWSHub
	defaultChatWSHub = hub
	defer func() {
		defaultChatWSHub = originalHub
	}()

	client := newTestChatWSClient(7, 1)
	hub.register(client)

	event := chatWSMessageEvent{
		MemberUserIDs: []uint{7},
		Message: map[string]any{
			"id": "101",
		},
	}
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	handleChatMessageEvent(context.Background(), string(raw))

	payload := receiveChatWSPayload(t, client.send)
	if payload["event"] != chatWSEventMessageCreated {
		t.Fatalf("unexpected event: %v", payload["event"])
	}
}

func TestChatWSHubSkipsOwnRedisMessageEvent(t *testing.T) {
	hub := newChatWSHub()
	originalHub := defaultChatWSHub
	defaultChatWSHub = hub
	defer func() {
		defaultChatWSHub = originalHub
	}()

	client := newTestChatWSClient(7, 1)
	hub.register(client)

	event := chatWSMessageEvent{
		Origin:        service.Chat.MessageEventOrigin(),
		MemberUserIDs: []uint{7},
		Message:       map[string]any{"id": "101"},
	}
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	handleChatMessageEvent(context.Background(), string(raw))

	select {
	case payload := <-client.send:
		t.Fatalf("own redis event should not be rebroadcast: %+v", payload)
	default:
	}
}

func TestSubscribeChatMessageEventsRetriesAfterFailure(t *testing.T) {
	originalSubscribe := chatWSSubscribeEvents
	originalWait := chatWSWaitSubscribeRetry
	defer func() {
		chatWSSubscribeEvents = originalSubscribe
		chatWSWaitSubscribeRetry = originalWait
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	attempts := 0
	chatWSSubscribeEvents = func(ctx context.Context) (<-chan string, func() error, error) {
		attempts++
		if attempts == 1 {
			return nil, nil, errors.New("redis unavailable")
		}
		ch := make(chan string)
		close(ch)
		cancel()
		return ch, func() error { return nil }, nil
	}
	chatWSWaitSubscribeRetry = func(ctx context.Context) bool {
		return ctx.Err() == nil
	}

	subscribeChatMessageEvents(ctx)
	if attempts != 2 {
		t.Fatalf("expected subscriber to retry once, got %d attempts", attempts)
	}
}

func TestPublishChatMessageEventWithRetryEventuallySucceeds(t *testing.T) {
	originalPublish := chatWSPublishMessageEvent
	originalWait := chatWSWaitPublishRetry
	defer func() {
		chatWSPublishMessageEvent = originalPublish
		chatWSWaitPublishRetry = originalWait
	}()

	attempts := 0
	chatWSPublishMessageEvent = func(ctx context.Context, memberIDs []uint, message *modelchat.ChatMessage) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary publish failure")
		}
		return nil
	}
	chatWSWaitPublishRetry = func(ctx context.Context) bool { return true }

	if err := publishChatMessageEventWithRetry(context.Background(), []uint{1}, &modelchat.ChatMessage{Id: "1"}); err != nil {
		t.Fatalf("expected retry success, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 publish attempts, got %d", attempts)
	}
}

func TestPublishChatMessageEventWithRetryReturnsLastError(t *testing.T) {
	originalPublish := chatWSPublishMessageEvent
	originalWait := chatWSWaitPublishRetry
	defer func() {
		chatWSPublishMessageEvent = originalPublish
		chatWSWaitPublishRetry = originalWait
	}()

	attempts := 0
	chatWSPublishMessageEvent = func(ctx context.Context, memberIDs []uint, message *modelchat.ChatMessage) error {
		attempts++
		return errors.New("publish failed")
	}
	chatWSWaitPublishRetry = func(ctx context.Context) bool { return true }

	if err := publishChatMessageEventWithRetry(context.Background(), []uint{1}, &modelchat.ChatMessage{Id: "1"}); err == nil {
		t.Fatal("expected publish retry to return error")
	}
	if attempts != chatWSPublishRetries {
		t.Fatalf("expected %d publish attempts, got %d", chatWSPublishRetries, attempts)
	}
}

func newTestChatWSClient(userID uint, bufferSize int) *chatWSClient {
	return &chatWSClient{
		userID: userID,
		send:   make(chan any, bufferSize),
	}
}

func receiveChatWSPayload(t *testing.T, ch <-chan any) map[string]any {
	t.Helper()

	select {
	case raw := <-ch:
		payload, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("unexpected payload type: %T", raw)
		}
		return payload
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for websocket payload")
	}
	return nil
}

func assertChannelClosed(t *testing.T, ch <-chan any) {
	t.Helper()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for closed channel")
	}
}

func assertChannelDrainsAndCloses(t *testing.T, ch <-chan any) {
	t.Helper()

	deadline := time.After(time.Second)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("timed out waiting for drained closed channel")
		}
	}
}

func assertNotPanics(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()
	fn()
}
