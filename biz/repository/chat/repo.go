package chat

import (
	"context"
	dbdal "video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	rdbdal "video-platform/biz/dal/rdb"
	userrepo "video-platform/biz/repository/user"
)

type chatDBStore interface {
	CreateRoom(ctx context.Context, room *model.ChatRoom, members []model.ChatRoomMember) error
	ListRoomsByUserID(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.ChatRoomListItem, error)
	GetRoomByID(ctx context.Context, roomID uint) (*model.ChatRoom, error)
	GetRoomMember(ctx context.Context, roomID, userID uint) (*model.ChatRoomMember, error)
	ListRoomMembers(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatRoomMember, error)
	ListRoomMessages(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatMessage, error)
	ListMessagesByIDs(ctx context.Context, messageIDs []uint) ([]model.ChatMessage, error)
	AddRoomMembers(ctx context.Context, members []model.ChatRoomMember) error
	DeleteRoomMember(ctx context.Context, roomID, userID uint) (bool, error)
	UpdateLastReadMessageID(ctx context.Context, roomID, userID, messageID uint) error
	CreateMessage(ctx context.Context, message *model.ChatMessage) error
}

type chatSnapshotStore interface {
	ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error)
}

type chatCacheStore interface {
	GetChatMessageCacheVersion(ctx context.Context, roomID uint) (int64, error)
	GetChatMessageCache(ctx context.Context, roomID uint, version int64, cursor uint, limit int, dest any) (bool, error)
	SetChatMessageCache(ctx context.Context, roomID uint, version int64, cursor uint, limit int, value any) error
	BumpChatMessageCacheVersion(ctx context.Context, roomID uint) error
}

type chatStore struct {
	db        chatDBStore
	snapshots chatSnapshotStore
	cache     chatCacheStore
}

type ChatRoomListResult struct {
	Rooms      []ChatRoomItem
	NextCursor uint
	HasMore    bool
}

type ChatRoomItem struct {
	Room        model.ChatRoom
	LastMessage *model.ChatMessage
	UnreadCount int64
	MemberCount int64
	Sender      *userrepo.UserProfile
}

type ChatMessageListResult struct {
	Messages   []ChatMessageItem
	NextCursor uint
	HasMore    bool
}

type ChatMessageItem struct {
	Message model.ChatMessage
	Sender  userrepo.UserProfile
}

type ChatRoomMemberListResult struct {
	Members    []ChatRoomMemberItem
	NextCursor uint
	HasMore    bool
}

type ChatRoomMemberItem struct {
	Member model.ChatRoomMember
	User   userrepo.UserProfile
}

type defaultChatSnapshotStore struct{}

func (defaultChatSnapshotStore) ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
	return userrepo.ListUserSnapshotsByIDs(ctx, userIDs)
}

var chats = chatStore{
	db:        dbdal.Chats,
	snapshots: defaultChatSnapshotStore{},
	cache:     rdbdal.Chats,
}

func CreateChatRoom(ctx context.Context, room *model.ChatRoom, members []model.ChatRoomMember) error {
	return chats.CreateRoom(ctx, room, members)
}

func ListChatRooms(ctx context.Context, userID uint, cursor uint, limit int) (*ChatRoomListResult, error) {
	return chats.ListRooms(ctx, userID, cursor, limit)
}

func GetChatRoomByID(ctx context.Context, roomID uint) (*model.ChatRoom, error) {
	return chats.db.GetRoomByID(ctx, roomID)
}

func GetChatRoomMember(ctx context.Context, roomID, userID uint) (*model.ChatRoomMember, error) {
	return chats.db.GetRoomMember(ctx, roomID, userID)
}

func ListChatRoomMembers(ctx context.Context, roomID uint, cursor uint, limit int) (*ChatRoomMemberListResult, error) {
	return chats.ListMembers(ctx, roomID, cursor, limit)
}

func ListChatMessages(ctx context.Context, roomID uint, cursor uint, limit int) (*ChatMessageListResult, error) {
	return chats.ListMessages(ctx, roomID, cursor, limit)
}

func AddChatRoomMembers(ctx context.Context, members []model.ChatRoomMember) error {
	return chats.db.AddRoomMembers(ctx, members)
}

func DeleteChatRoomMember(ctx context.Context, roomID, userID uint) (bool, error) {
	return chats.db.DeleteRoomMember(ctx, roomID, userID)
}

func UpdateChatLastReadMessageID(ctx context.Context, roomID, userID, messageID uint) error {
	return chats.db.UpdateLastReadMessageID(ctx, roomID, userID, messageID)
}

func CreateChatMessage(ctx context.Context, message *model.ChatMessage) error {
	return chats.CreateMessage(ctx, message)
}

func (s chatStore) CreateRoom(ctx context.Context, room *model.ChatRoom, members []model.ChatRoomMember) error {
	return s.db.CreateRoom(ctx, room, members)
}

func (s chatStore) CreateMessage(ctx context.Context, message *model.ChatMessage) error {
	if err := s.db.CreateMessage(ctx, message); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.BumpChatMessageCacheVersion(ctx, message.RoomID)
	}
	return nil
}

func (s chatStore) ListRooms(ctx context.Context, userID uint, cursor uint, limit int) (*ChatRoomListResult, error) {
	items, err := s.db.ListRoomsByUserID(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	nextCursor := uint(0)
	if hasMore && len(items) > 0 {
		nextCursor = items[len(items)-1].Room.ID
	}

	lastMessageIDs := make([]uint, 0, len(items))
	for _, item := range items {
		if item.Room.LastMessageID > 0 {
			lastMessageIDs = append(lastMessageIDs, item.Room.LastMessageID)
		}
	}

	lastMessages, err := s.db.ListMessagesByIDs(ctx, lastMessageIDs)
	if err != nil {
		return nil, err
	}

	messageMap := make(map[uint]model.ChatMessage, len(lastMessages))
	senderIDs := make([]uint, 0, len(lastMessages))
	for _, message := range lastMessages {
		messageMap[message.ID] = message
		senderIDs = append(senderIDs, message.SenderID)
	}

	senderMap, err := s.userMap(ctx, senderIDs)
	if err != nil {
		return nil, err
	}

	rooms := make([]ChatRoomItem, 0, len(items))
	for _, item := range items {
		roomItem := ChatRoomItem{
			Room:        item.Room,
			UnreadCount: item.UnreadCount,
			MemberCount: item.MemberCount,
		}
		if item.Room.LastMessageID > 0 {
			if lastMessage, ok := messageMap[item.Room.LastMessageID]; ok {
				roomItem.LastMessage = &lastMessage
				if sender, ok := senderMap[lastMessage.SenderID]; ok {
					roomItem.Sender = &sender
				}
			}
		}
		rooms = append(rooms, roomItem)
	}

	return &ChatRoomListResult{
		Rooms:      rooms,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s chatStore) ListMembers(ctx context.Context, roomID uint, cursor uint, limit int) (*ChatRoomMemberListResult, error) {
	members, err := s.db.ListRoomMembers(ctx, roomID, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(members) > limit
	if hasMore {
		members = members[:limit]
	}

	nextCursor := uint(0)
	if hasMore && len(members) > 0 {
		nextCursor = members[len(members)-1].ID
	}

	userIDs := make([]uint, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}

	userMap, err := s.userMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	items := make([]ChatRoomMemberItem, 0, len(members))
	for _, member := range members {
		items = append(items, ChatRoomMemberItem{
			Member: member,
			User:   userMap[member.UserID],
		})
	}

	return &ChatRoomMemberListResult{
		Members:    items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s chatStore) ListMessages(ctx context.Context, roomID uint, cursor uint, limit int) (*ChatMessageListResult, error) {
	version := int64(0)
	cacheEnabled := s.cache != nil && cursor > 0
	var result ChatMessageListResult
	if cacheEnabled {
		var err error
		version, err = s.cache.GetChatMessageCacheVersion(ctx, roomID)
		if err != nil {
			cacheEnabled = false
		}
	}

	if cacheEnabled {
		if ok, err := s.cache.GetChatMessageCache(ctx, roomID, version, cursor, limit, &result); err == nil && ok {
			return &result, nil
		}
	}

	messages, err := s.db.ListRoomMessages(ctx, roomID, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	nextCursor := uint(0)
	if hasMore && len(messages) > 0 {
		nextCursor = messages[len(messages)-1].ID
	}

	userIDs := make([]uint, 0, len(messages))
	for _, message := range messages {
		userIDs = append(userIDs, message.SenderID)
	}

	userMap, err := s.userMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	items := make([]ChatMessageItem, 0, len(messages))
	for _, message := range messages {
		items = append(items, ChatMessageItem{
			Message: message,
			Sender:  userMap[message.SenderID],
		})
	}

	result = ChatMessageListResult{
		Messages:   items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
	if cacheEnabled {
		_ = s.cache.SetChatMessageCache(ctx, roomID, version, cursor, limit, result)
	}

	return &result, nil
}

func (s chatStore) userMap(ctx context.Context, userIDs []uint) (map[uint]userrepo.UserProfile, error) {
	users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	userMap := make(map[uint]userrepo.UserProfile, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}
	return userMap, nil
}
