package chat

import (
	"context"
	"errors"
	"testing"
	dbdal "video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	userrepo "video-platform/biz/repository/user"
)

type fakeChatDBStore struct {
	createRoomFn              func(ctx context.Context, room *model.ChatRoom, members []model.ChatRoomMember) error
	listRoomsByUserIDFn       func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.ChatRoomListItem, error)
	getRoomByIDFn             func(ctx context.Context, roomID uint) (*model.ChatRoom, error)
	getRoomMemberFn           func(ctx context.Context, roomID, userID uint) (*model.ChatRoomMember, error)
	listRoomMembersFn         func(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatRoomMember, error)
	listRoomMessagesFn        func(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatMessage, error)
	listMessagesByIDsFn       func(ctx context.Context, messageIDs []uint) ([]model.ChatMessage, error)
	addRoomMembersFn          func(ctx context.Context, members []model.ChatRoomMember) error
	deleteRoomMemberFn        func(ctx context.Context, roomID, userID uint) (bool, error)
	updateLastReadMessageIDFn func(ctx context.Context, roomID, userID, messageID uint) error
	createMessageFn           func(ctx context.Context, message *model.ChatMessage) error
}

func (f fakeChatDBStore) CreateRoom(ctx context.Context, room *model.ChatRoom, members []model.ChatRoomMember) error {
	return f.createRoomFn(ctx, room, members)
}

func (f fakeChatDBStore) ListRoomsByUserID(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.ChatRoomListItem, error) {
	return f.listRoomsByUserIDFn(ctx, userID, cursor, limit)
}

func (f fakeChatDBStore) GetRoomByID(ctx context.Context, roomID uint) (*model.ChatRoom, error) {
	return f.getRoomByIDFn(ctx, roomID)
}

func (f fakeChatDBStore) GetRoomMember(ctx context.Context, roomID, userID uint) (*model.ChatRoomMember, error) {
	return f.getRoomMemberFn(ctx, roomID, userID)
}

func (f fakeChatDBStore) ListRoomMembers(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatRoomMember, error) {
	return f.listRoomMembersFn(ctx, roomID, cursor, limit)
}

func (f fakeChatDBStore) ListRoomMessages(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatMessage, error) {
	return f.listRoomMessagesFn(ctx, roomID, cursor, limit)
}

func (f fakeChatDBStore) ListMessagesByIDs(ctx context.Context, messageIDs []uint) ([]model.ChatMessage, error) {
	return f.listMessagesByIDsFn(ctx, messageIDs)
}

func (f fakeChatDBStore) AddRoomMembers(ctx context.Context, members []model.ChatRoomMember) error {
	return f.addRoomMembersFn(ctx, members)
}

func (f fakeChatDBStore) DeleteRoomMember(ctx context.Context, roomID, userID uint) (bool, error) {
	return f.deleteRoomMemberFn(ctx, roomID, userID)
}

func (f fakeChatDBStore) UpdateLastReadMessageID(ctx context.Context, roomID, userID, messageID uint) error {
	return f.updateLastReadMessageIDFn(ctx, roomID, userID, messageID)
}

func (f fakeChatDBStore) CreateMessage(ctx context.Context, message *model.ChatMessage) error {
	return f.createMessageFn(ctx, message)
}

type fakeChatSnapshotStore struct {
	listUserSnapshotsByIDsFn func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error)
}

func (f fakeChatSnapshotStore) ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
	return f.listUserSnapshotsByIDsFn(ctx, userIDs)
}

type fakeChatCacheStore struct {
	getChatMessageCacheVersionFn  func(ctx context.Context, roomID uint) (int64, error)
	getChatMessageCacheFn         func(ctx context.Context, roomID uint, version int64, cursor uint, limit int, dest any) (bool, error)
	setChatMessageCacheFn         func(ctx context.Context, roomID uint, version int64, cursor uint, limit int, value any) error
	bumpChatMessageCacheVersionFn func(ctx context.Context, roomID uint) error
}

func (f fakeChatCacheStore) GetChatMessageCacheVersion(ctx context.Context, roomID uint) (int64, error) {
	return f.getChatMessageCacheVersionFn(ctx, roomID)
}

func (f fakeChatCacheStore) GetChatMessageCache(ctx context.Context, roomID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
	return f.getChatMessageCacheFn(ctx, roomID, version, cursor, limit, dest)
}

func (f fakeChatCacheStore) SetChatMessageCache(ctx context.Context, roomID uint, version int64, cursor uint, limit int, value any) error {
	return f.setChatMessageCacheFn(ctx, roomID, version, cursor, limit, value)
}

func (f fakeChatCacheStore) BumpChatMessageCacheVersion(ctx context.Context, roomID uint) error {
	return f.bumpChatMessageCacheVersionFn(ctx, roomID)
}

func TestChatStoreListRoomsBuildsRoomItems(t *testing.T) {
	store := chatStore{
		db: fakeChatDBStore{
			listRoomsByUserIDFn: func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.ChatRoomListItem, error) {
				if userID != 7 || cursor != 0 || limit != 2 {
					t.Fatalf("unexpected list params userID=%d cursor=%d limit=%d", userID, cursor, limit)
				}
				return []dbdal.ChatRoomListItem{
					{Room: model.ChatRoom{ID: 11, LastMessageID: 101}, UnreadCount: 3, MemberCount: 2},
					{Room: model.ChatRoom{ID: 10, LastMessageID: 100}, UnreadCount: 1, MemberCount: 4},
				}, nil
			},
			listMessagesByIDsFn: func(ctx context.Context, messageIDs []uint) ([]model.ChatMessage, error) {
				if len(messageIDs) != 1 || messageIDs[0] != 101 {
					t.Fatalf("unexpected message IDs: %+v", messageIDs)
				}
				return []model.ChatMessage{
					{ID: 101, SenderID: 8, Content: "latest"},
				}, nil
			},
		},
		snapshots: fakeChatSnapshotStore{
			listUserSnapshotsByIDsFn: func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
				if len(userIDs) != 1 || userIDs[0] != 8 {
					t.Fatalf("unexpected snapshot IDs: %+v", userIDs)
				}
				return []userrepo.UserProfile{
					{ID: 8, Username: "alice"},
				}, nil
			},
		},
	}

	got, err := store.ListRooms(context.Background(), 7, 0, 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(got.Rooms))
	}
	if !got.HasMore || got.NextCursor != 11 {
		t.Fatalf("unexpected pagination: %+v", got)
	}
	if got.Rooms[0].LastMessage == nil || got.Rooms[0].LastMessage.Content != "latest" {
		t.Fatalf("unexpected last message: %+v", got.Rooms[0].LastMessage)
	}
	if got.Rooms[0].Sender == nil || got.Rooms[0].Sender.Username != "alice" {
		t.Fatalf("unexpected sender: %+v", got.Rooms[0].Sender)
	}
}

func TestChatStoreListMembersBuildsMemberItems(t *testing.T) {
	store := chatStore{
		db: fakeChatDBStore{
			listRoomMembersFn: func(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatRoomMember, error) {
				return []model.ChatRoomMember{
					{ID: 3, RoomID: roomID, UserID: 12},
					{ID: 2, RoomID: roomID, UserID: 13},
				}, nil
			},
		},
		snapshots: fakeChatSnapshotStore{
			listUserSnapshotsByIDsFn: func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
				return []userrepo.UserProfile{
					{ID: 12, Username: "tom"},
					{ID: 13, Username: "jerry"},
				}, nil
			},
		},
	}

	got, err := store.ListMembers(context.Background(), 5, 0, 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Members) != 1 || got.Members[0].User.Username != "tom" {
		t.Fatalf("unexpected members: %+v", got.Members)
	}
	if !got.HasMore || got.NextCursor != 3 {
		t.Fatalf("unexpected pagination: %+v", got)
	}
}

func TestChatStoreListMessagesBuildsMessageItems(t *testing.T) {
	var cachedValue ChatMessageListResult

	store := chatStore{
		db: fakeChatDBStore{
			listRoomMessagesFn: func(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatMessage, error) {
				return []model.ChatMessage{
					{ID: 21, RoomID: roomID, SenderID: 1, Content: "hello"},
					{ID: 20, RoomID: roomID, SenderID: 2, Content: "world"},
				}, nil
			},
		},
		cache: fakeChatCacheStore{
			getChatMessageCacheVersionFn: func(ctx context.Context, roomID uint) (int64, error) {
				return 2, nil
			},
			getChatMessageCacheFn: func(ctx context.Context, roomID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
				if roomID != 6 || version != 2 || cursor != 30 || limit != 1 {
					t.Fatalf("unexpected cache params roomID=%d version=%d cursor=%d limit=%d", roomID, version, cursor, limit)
				}
				return false, nil
			},
			setChatMessageCacheFn: func(ctx context.Context, roomID uint, version int64, cursor uint, limit int, value any) error {
				if roomID != 6 || version != 2 || cursor != 30 || limit != 1 {
					t.Fatalf("unexpected cache set params roomID=%d version=%d cursor=%d limit=%d", roomID, version, cursor, limit)
				}
				cachedValue = value.(ChatMessageListResult)
				return nil
			},
		},
		snapshots: fakeChatSnapshotStore{
			listUserSnapshotsByIDsFn: func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
				return []userrepo.UserProfile{
					{ID: 1, Username: "alice"},
					{ID: 2, Username: "bob"},
				}, nil
			},
		},
	}

	got, err := store.ListMessages(context.Background(), 6, 30, 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Messages) != 1 || got.Messages[0].Sender.Username != "alice" {
		t.Fatalf("unexpected messages: %+v", got.Messages)
	}
	if !got.HasMore || got.NextCursor != 21 {
		t.Fatalf("unexpected pagination: %+v", got)
	}
	if len(cachedValue.Messages) != 1 || cachedValue.Messages[0].Message.ID != 21 {
		t.Fatalf("unexpected cached value: %+v", cachedValue)
	}
}

func TestChatStoreListMessagesUsesCache(t *testing.T) {
	store := chatStore{
		db: fakeChatDBStore{
			listRoomMessagesFn: func(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatMessage, error) {
				t.Fatal("db should not be called on cache hit")
				return nil, nil
			},
		},
		cache: fakeChatCacheStore{
			getChatMessageCacheVersionFn: func(ctx context.Context, roomID uint) (int64, error) {
				return 3, nil
			},
			getChatMessageCacheFn: func(ctx context.Context, roomID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
				result := dest.(*ChatMessageListResult)
				result.Messages = []ChatMessageItem{
					{Message: model.ChatMessage{ID: 31, RoomID: roomID, SenderID: 1, Content: "cached"}},
				}
				result.HasMore = false
				return true, nil
			},
		},
	}

	got, err := store.ListMessages(context.Background(), 6, 30, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Messages) != 1 || got.Messages[0].Message.Content != "cached" {
		t.Fatalf("unexpected cached messages: %+v", got.Messages)
	}
}

func TestChatStoreListLatestMessagesSkipsCache(t *testing.T) {
	store := chatStore{
		db: fakeChatDBStore{
			listRoomMessagesFn: func(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatMessage, error) {
				if cursor != 0 {
					t.Fatalf("expected latest cursor 0, got %d", cursor)
				}
				return []model.ChatMessage{
					{ID: 21, RoomID: roomID, SenderID: 1, Content: "latest"},
				}, nil
			},
		},
		cache: fakeChatCacheStore{
			getChatMessageCacheVersionFn: func(ctx context.Context, roomID uint) (int64, error) {
				t.Fatal("latest page should not read cache version")
				return 0, nil
			},
			getChatMessageCacheFn: func(ctx context.Context, roomID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
				t.Fatal("latest page should not read cache")
				return false, nil
			},
			setChatMessageCacheFn: func(ctx context.Context, roomID uint, version int64, cursor uint, limit int, value any) error {
				t.Fatal("latest page should not write cache")
				return nil
			},
		},
		snapshots: fakeChatSnapshotStore{
			listUserSnapshotsByIDsFn: func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
				return []userrepo.UserProfile{{ID: 1, Username: "alice"}}, nil
			},
		},
	}

	got, err := store.ListMessages(context.Background(), 6, 0, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Messages) != 1 || got.Messages[0].Message.Content != "latest" {
		t.Fatalf("unexpected messages: %+v", got.Messages)
	}
}

func TestChatStoreListMessagesFallsBackToDBOnVersionCacheError(t *testing.T) {
	wantErr := errors.New("redis unavailable")

	store := chatStore{
		db: fakeChatDBStore{
			listRoomMessagesFn: func(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatMessage, error) {
				return []model.ChatMessage{
					{ID: 21, RoomID: roomID, SenderID: 1, Content: "db"},
				}, nil
			},
		},
		cache: fakeChatCacheStore{
			getChatMessageCacheVersionFn: func(ctx context.Context, roomID uint) (int64, error) {
				return 0, wantErr
			},
			getChatMessageCacheFn: func(ctx context.Context, roomID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
				t.Fatal("message cache should not be read when version lookup fails")
				return false, nil
			},
			setChatMessageCacheFn: func(ctx context.Context, roomID uint, version int64, cursor uint, limit int, value any) error {
				t.Fatal("message cache should not be written when version lookup fails")
				return nil
			},
		},
		snapshots: fakeChatSnapshotStore{
			listUserSnapshotsByIDsFn: func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
				return []userrepo.UserProfile{{ID: 1, Username: "alice"}}, nil
			},
		},
	}

	got, err := store.ListMessages(context.Background(), 6, 30, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Messages) != 1 || got.Messages[0].Message.Content != "db" {
		t.Fatalf("unexpected messages: %+v", got.Messages)
	}
}

func TestChatStoreCreateMessageBumpsMessageCacheVersion(t *testing.T) {
	var bumpedRoomID uint

	store := chatStore{
		db: fakeChatDBStore{
			createMessageFn: func(ctx context.Context, message *model.ChatMessage) error {
				if message.RoomID != 6 {
					t.Fatalf("expected roomID 6, got %d", message.RoomID)
				}
				return nil
			},
		},
		cache: fakeChatCacheStore{
			bumpChatMessageCacheVersionFn: func(ctx context.Context, roomID uint) error {
				bumpedRoomID = roomID
				return nil
			},
		},
	}

	if err := store.CreateMessage(context.Background(), &model.ChatMessage{RoomID: 6}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if bumpedRoomID != 6 {
		t.Fatalf("expected bumped roomID 6, got %d", bumpedRoomID)
	}
}

func TestChatStoreListRoomsReturnsMessageError(t *testing.T) {
	wantErr := errors.New("message lookup failed")
	store := chatStore{
		db: fakeChatDBStore{
			listRoomsByUserIDFn: func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.ChatRoomListItem, error) {
				return []dbdal.ChatRoomListItem{{Room: model.ChatRoom{ID: 1, LastMessageID: 99}}}, nil
			},
			listMessagesByIDsFn: func(ctx context.Context, messageIDs []uint) ([]model.ChatMessage, error) {
				return nil, wantErr
			},
		},
		snapshots: fakeChatSnapshotStore{
			listUserSnapshotsByIDsFn: func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
				t.Fatal("snapshots should not be called on message error")
				return nil, nil
			},
		},
	}

	_, err := store.ListRooms(context.Background(), 1, 0, 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
