package chat

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"video-platform/biz/dal/model"
	chat "video-platform/biz/model/chat"
	chatrepo "video-platform/biz/repository/chat"
	userrepo "video-platform/biz/repository/user"
	interactionsvc "video-platform/biz/service/interaction"
	usersvc "video-platform/biz/service/user"
	"video-platform/pkg/pagination"

	"gorm.io/gorm"
)

const (
	chatRoomTypePrivate = "private"
	chatRoomTypeGroup   = "group"

	chatMemberRoleOwner  = "owner"
	chatMemberRoleMember = "member"

	chatMessageTypeText     = "text"
	chatMessageStatusNormal = "normal"
)

type chatRepository interface {
	GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error)
	CreateChatRoom(ctx context.Context, room *model.ChatRoom, members []model.ChatRoomMember) error
	ListChatRooms(ctx context.Context, userID uint, cursor uint, limit int) (*chatrepo.ChatRoomListResult, error)
	GetChatRoomByID(ctx context.Context, roomID uint) (*model.ChatRoom, error)
	GetChatRoomMember(ctx context.Context, roomID, userID uint) (*model.ChatRoomMember, error)
	ListChatRoomMembers(ctx context.Context, roomID uint, cursor uint, limit int) (*chatrepo.ChatRoomMemberListResult, error)
	ListChatMessages(ctx context.Context, roomID uint, cursor uint, limit int) (*chatrepo.ChatMessageListResult, error)
	AddChatRoomMembers(ctx context.Context, members []model.ChatRoomMember) error
	DeleteChatRoomMember(ctx context.Context, roomID, userID uint) (bool, error)
	UpdateChatLastReadMessageID(ctx context.Context, roomID, userID, messageID uint) error
	CreateChatMessage(ctx context.Context, message *model.ChatMessage) error
	PublishChatMessageEvent(ctx context.Context, event chatrepo.ChatMessageEvent) error
	SubscribeChatMessageEvents(ctx context.Context) (<-chan string, func() error, error)
}

type defaultChatRepository struct{}

func (defaultChatRepository) GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
	return userrepo.GetUserByID(ctx, userID)
}

func (defaultChatRepository) CreateChatRoom(ctx context.Context, room *model.ChatRoom, members []model.ChatRoomMember) error {
	return chatrepo.CreateChatRoom(ctx, room, members)
}

func (defaultChatRepository) ListChatRooms(ctx context.Context, userID uint, cursor uint, limit int) (*chatrepo.ChatRoomListResult, error) {
	return chatrepo.ListChatRooms(ctx, userID, cursor, limit)
}

func (defaultChatRepository) GetChatRoomByID(ctx context.Context, roomID uint) (*model.ChatRoom, error) {
	return chatrepo.GetChatRoomByID(ctx, roomID)
}

func (defaultChatRepository) GetChatRoomMember(ctx context.Context, roomID, userID uint) (*model.ChatRoomMember, error) {
	return chatrepo.GetChatRoomMember(ctx, roomID, userID)
}

func (defaultChatRepository) ListChatRoomMembers(ctx context.Context, roomID uint, cursor uint, limit int) (*chatrepo.ChatRoomMemberListResult, error) {
	return chatrepo.ListChatRoomMembers(ctx, roomID, cursor, limit)
}

func (defaultChatRepository) ListChatMessages(ctx context.Context, roomID uint, cursor uint, limit int) (*chatrepo.ChatMessageListResult, error) {
	return chatrepo.ListChatMessages(ctx, roomID, cursor, limit)
}

func (defaultChatRepository) AddChatRoomMembers(ctx context.Context, members []model.ChatRoomMember) error {
	return chatrepo.AddChatRoomMembers(ctx, members)
}

func (defaultChatRepository) DeleteChatRoomMember(ctx context.Context, roomID, userID uint) (bool, error) {
	return chatrepo.DeleteChatRoomMember(ctx, roomID, userID)
}

func (defaultChatRepository) UpdateChatLastReadMessageID(ctx context.Context, roomID, userID, messageID uint) error {
	return chatrepo.UpdateChatLastReadMessageID(ctx, roomID, userID, messageID)
}

func (defaultChatRepository) CreateChatMessage(ctx context.Context, message *model.ChatMessage) error {
	return chatrepo.CreateChatMessage(ctx, message)
}

func (defaultChatRepository) PublishChatMessageEvent(ctx context.Context, event chatrepo.ChatMessageEvent) error {
	return chatrepo.PublishChatMessageEvent(ctx, event)
}

func (defaultChatRepository) SubscribeChatMessageEvents(ctx context.Context) (<-chan string, func() error, error) {
	return chatrepo.SubscribeChatMessageEvents(ctx)
}

func (s chatService) ListMemberUserIDs(ctx context.Context, userID, roomID uint) ([]uint, error) {
	if err := s.ensureRoomMember(ctx, roomID, userID); err != nil {
		return nil, err
	}

	const memberBatchSize = 100
	cursor := uint(0)
	memberIDs := make([]uint, 0)
	for {
		result, err := s.repo.ListChatRoomMembers(ctx, roomID, cursor, memberBatchSize)
		if err != nil {
			return nil, err
		}
		for _, item := range result.Members {
			memberIDs = append(memberIDs, item.Member.UserID)
		}
		if !result.HasMore {
			break
		}
		cursor = result.NextCursor
	}
	return memberIDs, nil
}

func (s chatService) PublishMessageEvent(ctx context.Context, memberUserIDs []uint, message *chat.ChatMessage) error {
	return s.repo.PublishChatMessageEvent(ctx, chatrepo.ChatMessageEvent{
		MemberUserIDs: memberUserIDs,
		Message:       message,
	})
}

func (s chatService) SubscribeMessageEvents(ctx context.Context) (<-chan string, func() error, error) {
	return s.repo.SubscribeChatMessageEvents(ctx)
}

func (s chatService) MessageEventOrigin() string {
	return chatrepo.ChatMessageEventOrigin()
}

type chatService struct {
	repo chatRepository
}

var Chat = chatService{
	repo: defaultChatRepository{},
}

func (s chatService) CreateRoom(ctx context.Context, ownerID uint, req *chat.CreateChatRoomRequest) (*chat.ChatRoom, error) {
	if _, err := s.repo.GetUserByID(ctx, ownerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usersvc.ErrUserNotFound
		}
		return nil, err
	}

	roomType, err := mapChatRoomType(req.Type)
	if err != nil {
		return nil, err
	}

	memberIDs, err := normalizeChatMemberIDs(ownerID, req.MemberUserIds)
	if err != nil {
		return nil, err
	}
	if roomType == chatRoomTypePrivate && len(memberIDs) != 2 {
		return nil, ErrChatPrivateMemberCount
	}
	if roomType == chatRoomTypeGroup && len(memberIDs) < 2 {
		return nil, ErrChatMemberRequired
	}
	if roomType == chatRoomTypeGroup && strings.TrimSpace(req.Name) == "" {
		return nil, ErrChatGroupNameRequired
	}

	for _, memberID := range memberIDs {
		if _, err := s.repo.GetUserByID(ctx, memberID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, usersvc.ErrUserNotFound
			}
			return nil, err
		}
	}

	room := &model.ChatRoom{
		Type:    roomType,
		Name:    strings.TrimSpace(req.Name),
		OwnerID: ownerID,
	}
	if roomType == chatRoomTypePrivate {
		room.Name = ""
		room.OwnerID = 0
	}

	now := time.Now()
	members := make([]model.ChatRoomMember, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		role := chatMemberRoleMember
		if roomType == chatRoomTypeGroup && memberID == ownerID {
			role = chatMemberRoleOwner
		}
		members = append(members, model.ChatRoomMember{
			UserID:   memberID,
			Role:     role,
			JoinedAt: now,
		})
	}

	if err := s.repo.CreateChatRoom(ctx, room, members); err != nil {
		return nil, err
	}

	return buildChatRoom(chatrepo.ChatRoomItem{Room: *room}), nil
}

func (s chatService) ListRooms(ctx context.Context, userID uint, cursor uint, limit int32) (*chat.ChatRoomList, error) {
	result, err := s.repo.ListChatRooms(ctx, userID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}
	return buildChatRoomList(result), nil
}

func (s chatService) ListMessages(ctx context.Context, userID, roomID uint, cursor uint, limit int32) (*chat.ChatMessageList, error) {
	if err := s.ensureRoomMember(ctx, roomID, userID); err != nil {
		return nil, err
	}

	result, err := s.repo.ListChatMessages(ctx, roomID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}
	return buildChatMessageList(result), nil
}

func (s chatService) ListMembers(ctx context.Context, userID, roomID uint, cursor uint, limit int32) (*chat.ChatRoomMemberList, error) {
	if err := s.ensureRoomMember(ctx, roomID, userID); err != nil {
		return nil, err
	}

	result, err := s.repo.ListChatRoomMembers(ctx, roomID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}
	return buildChatRoomMemberList(result), nil
}

func (s chatService) InviteMembers(ctx context.Context, userID, roomID uint, rawMemberIDs []string) error {
	room, err := s.repo.GetChatRoomByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatRoomNotFound
		}
		return err
	}
	if room.Type != chatRoomTypeGroup {
		return interactionsvc.ErrNoPermission
	}
	operator, err := s.repo.GetChatRoomMember(ctx, roomID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatRoomMemberNotFound
		}
		return err
	}
	if operator.Role != chatMemberRoleOwner && operator.Role != "admin" {
		return interactionsvc.ErrNoPermission
	}

	memberIDs, err := normalizeChatMemberIDs(0, rawMemberIDs)
	if err != nil {
		return err
	}
	if len(memberIDs) == 0 {
		return ErrChatMemberRequired
	}

	now := time.Now()
	members := make([]model.ChatRoomMember, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		if _, err := s.repo.GetUserByID(ctx, memberID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return usersvc.ErrUserNotFound
			}
			return err
		}
		if _, err := s.repo.GetChatRoomMember(ctx, roomID, memberID); err == nil {
			return ErrChatMemberExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		members = append(members, model.ChatRoomMember{
			RoomID:   roomID,
			UserID:   memberID,
			Role:     chatMemberRoleMember,
			JoinedAt: now,
		})
	}

	return s.repo.AddChatRoomMembers(ctx, members)
}

func (s chatService) LeaveRoom(ctx context.Context, userID, roomID uint) error {
	member, err := s.repo.GetChatRoomMember(ctx, roomID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatRoomMemberNotFound
		}
		return err
	}
	if member.Role == chatMemberRoleOwner {
		return ErrChatOwnerCannotLeave
	}

	deleted, err := s.repo.DeleteChatRoomMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrChatRoomMemberNotFound
	}
	return nil
}

func (s chatService) MarkRoomRead(ctx context.Context, userID, roomID, messageID uint) error {
	if err := s.ensureRoomMember(ctx, roomID, userID); err != nil {
		return err
	}
	if err := s.repo.UpdateChatLastReadMessageID(ctx, roomID, userID, messageID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatMessageNotFound
		}
		return err
	}
	return nil
}

func (s chatService) CreateMessage(ctx context.Context, senderID, roomID uint, content, clientMsgID string) (*chat.ChatMessage, error) {
	if err := s.ensureRoomMember(ctx, roomID, senderID); err != nil {
		return nil, err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, interactionsvc.ErrCommentEmpty
	}

	message := &model.ChatMessage{
		RoomID:      roomID,
		SenderID:    senderID,
		MessageType: chatMessageTypeText,
		Content:     content,
		Status:      chatMessageStatusNormal,
		ClientMsgID: clientMsgID,
	}
	if err := s.repo.CreateChatMessage(ctx, message); err != nil {
		return nil, err
	}

	sender, err := s.repo.GetUserByID(ctx, senderID)
	if err != nil {
		return nil, err
	}
	return buildChatMessage(chatrepo.ChatMessageItem{Message: *message, Sender: *sender}), nil
}

func (s chatService) ensureRoomMember(ctx context.Context, roomID, userID uint) error {
	if _, err := s.repo.GetChatRoomByID(ctx, roomID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatRoomNotFound
		}
		return err
	}
	if _, err := s.repo.GetChatRoomMember(ctx, roomID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatRoomMemberNotFound
		}
		return err
	}
	return nil
}

func normalizeChatMemberIDs(ownerID uint, rawIDs []string) ([]uint, error) {
	seen := make(map[uint]struct{}, len(rawIDs)+1)
	ids := make([]uint, 0, len(rawIDs)+1)
	if ownerID > 0 {
		seen[ownerID] = struct{}{}
		ids = append(ids, ownerID)
	}
	for _, rawID := range rawIDs {
		parsedID, err := strconv.ParseUint(rawID, 10, 64)
		if err != nil || parsedID == 0 {
			return nil, usersvc.ErrUserNotFound
		}
		id := uint(parsedID)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func mapChatRoomType(roomType chat.ChatRoomType) (string, error) {
	switch roomType {
	case chat.ChatRoomType_CHAT_ROOM_TYPE_PRIVATE:
		return chatRoomTypePrivate, nil
	case chat.ChatRoomType_CHAT_ROOM_TYPE_GROUP:
		return chatRoomTypeGroup, nil
	default:
		return "", ErrChatMemberRequired
	}
}

func mapChatRoomTypeToPB(roomType string) chat.ChatRoomType {
	switch roomType {
	case chatRoomTypePrivate:
		return chat.ChatRoomType_CHAT_ROOM_TYPE_PRIVATE
	case chatRoomTypeGroup:
		return chat.ChatRoomType_CHAT_ROOM_TYPE_GROUP
	default:
		return chat.ChatRoomType_CHAT_ROOM_TYPE_UNSPECIFIED
	}
}

func mapChatMemberRoleToPB(role string) chat.ChatMemberRole {
	switch role {
	case chatMemberRoleOwner:
		return chat.ChatMemberRole_CHAT_MEMBER_ROLE_OWNER
	case "admin":
		return chat.ChatMemberRole_CHAT_MEMBER_ROLE_ADMIN
	case chatMemberRoleMember:
		return chat.ChatMemberRole_CHAT_MEMBER_ROLE_MEMBER
	default:
		return chat.ChatMemberRole_CHAT_MEMBER_ROLE_UNSPECIFIED
	}
}

func mapChatMessageTypeToPB(messageType string) chat.ChatMessageType {
	switch messageType {
	case chatMessageTypeText:
		return chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT
	case "image":
		return chat.ChatMessageType_CHAT_MESSAGE_TYPE_IMAGE
	case "video":
		return chat.ChatMessageType_CHAT_MESSAGE_TYPE_VIDEO
	case "system":
		return chat.ChatMessageType_CHAT_MESSAGE_TYPE_SYSTEM
	default:
		return chat.ChatMessageType_CHAT_MESSAGE_TYPE_UNSPECIFIED
	}
}

func mapChatMessageStatusToPB(status string) chat.ChatMessageStatus {
	switch status {
	case chatMessageStatusNormal:
		return chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_NORMAL
	case "recalled":
		return chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_RECALLED
	case "deleted":
		return chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_DELETED
	default:
		return chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_UNSPECIFIED
	}
}

func buildChatRoomList(result *chatrepo.ChatRoomListResult) *chat.ChatRoomList {
	items := make([]*chat.ChatRoom, 0)
	if result == nil {
		return &chat.ChatRoomList{Items: items}
	}
	items = make([]*chat.ChatRoom, 0, len(result.Rooms))
	for _, room := range result.Rooms {
		items = append(items, buildChatRoom(room))
	}

	nextCursor := ""
	if result.HasMore {
		nextCursor = strconv.FormatUint(uint64(result.NextCursor), 10)
	}
	return &chat.ChatRoomList{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	}
}

func buildChatRoom(item chatrepo.ChatRoomItem) *chat.ChatRoom {
	room := item.Room
	resp := &chat.ChatRoom{
		Id:          strconv.FormatUint(uint64(room.ID), 10),
		Type:        mapChatRoomTypeToPB(room.Type),
		Name:        room.Name,
		OwnerId:     formatOptionalUint(room.OwnerID),
		UnreadCount: item.UnreadCount,
		MemberCount: item.MemberCount,
		CreatedAt:   room.CreatedAt.Format(time.RFC3339),
	}
	if room.LastMessageAt != nil {
		resp.LastMessageAt = room.LastMessageAt.Format(time.RFC3339)
	}
	if item.LastMessage != nil {
		sender := userrepo.UserProfile{}
		if item.Sender != nil {
			sender = *item.Sender
		}
		resp.LastMessage = buildChatMessage(chatrepo.ChatMessageItem{
			Message: *item.LastMessage,
			Sender:  sender,
		})
	}
	return resp
}

func buildChatMessageList(result *chatrepo.ChatMessageListResult) *chat.ChatMessageList {
	items := make([]*chat.ChatMessage, 0)
	if result == nil {
		return &chat.ChatMessageList{Items: items}
	}
	items = make([]*chat.ChatMessage, 0, len(result.Messages))
	for _, message := range result.Messages {
		items = append(items, buildChatMessage(message))
	}

	nextCursor := ""
	if result.HasMore {
		nextCursor = strconv.FormatUint(uint64(result.NextCursor), 10)
	}
	return &chat.ChatMessageList{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	}
}

func buildChatMessage(item chatrepo.ChatMessageItem) *chat.ChatMessage {
	message := item.Message
	return &chat.ChatMessage{
		Id:              strconv.FormatUint(uint64(message.ID), 10),
		RoomId:          strconv.FormatUint(uint64(message.RoomID), 10),
		SenderId:        strconv.FormatUint(uint64(message.SenderID), 10),
		SenderUsername:  item.Sender.Username,
		SenderAvatarUrl: item.Sender.AvatarURL,
		MessageType:     mapChatMessageTypeToPB(message.MessageType),
		Content:         message.Content,
		Status:          mapChatMessageStatusToPB(message.Status),
		ClientMsgId:     message.ClientMsgID,
		CreatedAt:       message.CreatedAt.Format(time.RFC3339),
	}
}

func buildChatRoomMemberList(result *chatrepo.ChatRoomMemberListResult) *chat.ChatRoomMemberList {
	items := make([]*chat.ChatRoomMember, 0)
	if result == nil {
		return &chat.ChatRoomMemberList{Items: items}
	}
	items = make([]*chat.ChatRoomMember, 0, len(result.Members))
	for _, member := range result.Members {
		items = append(items, &chat.ChatRoomMember{
			RoomId:    strconv.FormatUint(uint64(member.Member.RoomID), 10),
			UserId:    strconv.FormatUint(uint64(member.Member.UserID), 10),
			Username:  member.User.Username,
			AvatarUrl: member.User.AvatarURL,
			Role:      mapChatMemberRoleToPB(member.Member.Role),
			JoinedAt:  member.Member.JoinedAt.Format(time.RFC3339),
		})
	}

	nextCursor := ""
	if result.HasMore {
		nextCursor = strconv.FormatUint(uint64(result.NextCursor), 10)
	}
	return &chat.ChatRoomMemberList{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	}
}

func formatOptionalUint(value uint) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatUint(uint64(value), 10)
}
