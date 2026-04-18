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
	store := chatStore{
		db: fakeChatDBStore{
			listRoomMessagesFn: func(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatMessage, error) {
				return []model.ChatMessage{
					{ID: 21, RoomID: roomID, SenderID: 1, Content: "hello"},
					{ID: 20, RoomID: roomID, SenderID: 2, Content: "world"},
				}, nil
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

	got, err := store.ListMessages(context.Background(), 6, 0, 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Messages) != 1 || got.Messages[0].Sender.Username != "alice" {
		t.Fatalf("unexpected messages: %+v", got.Messages)
	}
	if !got.HasMore || got.NextCursor != 21 {
		t.Fatalf("unexpected pagination: %+v", got)
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
