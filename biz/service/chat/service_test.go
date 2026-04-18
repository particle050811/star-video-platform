package chat

import (
	"context"
	"errors"
	"testing"
	"video-platform/biz/dal/model"
	chat "video-platform/biz/model/chat"
	chatrepo "video-platform/biz/repository/chat"
	userrepo "video-platform/biz/repository/user"
	interactionsvc "video-platform/biz/service/interaction"

	"gorm.io/gorm"
)

type fakeChatRepo struct {
	users         map[uint]*userrepo.UserProfile
	members       map[uint]map[uint]*model.ChatRoomMember
	rooms         map[uint]*model.ChatRoom
	createdRoom   *model.ChatRoom
	createdMember []model.ChatRoomMember
	updateReadErr error
	deletedMember bool
}

func (f *fakeChatRepo) GetUserByID(_ context.Context, userID uint) (*userrepo.UserProfile, error) {
	if user, ok := f.users[userID]; ok {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeChatRepo) CreateChatRoom(_ context.Context, room *model.ChatRoom, members []model.ChatRoomMember) error {
	room.ID = 10
	f.createdRoom = room
	f.createdMember = members
	return nil
}

func (f *fakeChatRepo) ListChatRooms(_ context.Context, _ uint, _ uint, _ int) (*chatrepo.ChatRoomListResult, error) {
	return &chatrepo.ChatRoomListResult{}, nil
}

func (f *fakeChatRepo) GetChatRoomByID(_ context.Context, roomID uint) (*model.ChatRoom, error) {
	if room, ok := f.rooms[roomID]; ok {
		return room, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeChatRepo) GetChatRoomMember(_ context.Context, roomID, userID uint) (*model.ChatRoomMember, error) {
	if roomMembers, ok := f.members[roomID]; ok {
		if member, ok := roomMembers[userID]; ok {
			return member, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeChatRepo) ListChatRoomMembers(_ context.Context, _ uint, _ uint, _ int) (*chatrepo.ChatRoomMemberListResult, error) {
	return &chatrepo.ChatRoomMemberListResult{}, nil
}

func (f *fakeChatRepo) ListChatMessages(_ context.Context, _ uint, _ uint, _ int) (*chatrepo.ChatMessageListResult, error) {
	return &chatrepo.ChatMessageListResult{}, nil
}

func (f *fakeChatRepo) AddChatRoomMembers(_ context.Context, _ []model.ChatRoomMember) error {
	return nil
}

func (f *fakeChatRepo) DeleteChatRoomMember(_ context.Context, _, _ uint) (bool, error) {
	return f.deletedMember, nil
}

func (f *fakeChatRepo) UpdateChatLastReadMessageID(_ context.Context, _, _, _ uint) error {
	return f.updateReadErr
}

func (f *fakeChatRepo) CreateChatMessage(_ context.Context, _ *model.ChatMessage) error {
	return nil
}

func TestChatServiceCreatePrivateRoomUsesMemberRole(t *testing.T) {
	repo := &fakeChatRepo{
		users: map[uint]*userrepo.UserProfile{
			1: {ID: 1, Username: "alice"},
			2: {ID: 2, Username: "bob"},
		},
	}
	svc := chatService{repo: repo}

	room, err := svc.CreateRoom(context.Background(), 1, &chat.CreateChatRoomRequest{
		Type:          chat.ChatRoomType_CHAT_ROOM_TYPE_PRIVATE,
		MemberUserIds: []string{"2"},
	})
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	if room.OwnerId != "" {
		t.Fatalf("expected private room owner_id empty, got %q", room.OwnerId)
	}
	if repo.createdRoom.OwnerID != 0 {
		t.Fatalf("expected private room owner id 0, got %d", repo.createdRoom.OwnerID)
	}
	for _, member := range repo.createdMember {
		if member.Role != chatMemberRoleMember {
			t.Fatalf("expected private member role %q, got %q", chatMemberRoleMember, member.Role)
		}
	}
}

func TestChatServiceCreatePrivateRoomRequiresTwoMembers(t *testing.T) {
	repo := &fakeChatRepo{
		users: map[uint]*userrepo.UserProfile{
			1: {ID: 1, Username: "alice"},
		},
	}
	svc := chatService{repo: repo}

	_, err := svc.CreateRoom(context.Background(), 1, &chat.CreateChatRoomRequest{
		Type: chat.ChatRoomType_CHAT_ROOM_TYPE_PRIVATE,
	})
	if !errors.Is(err, ErrChatPrivateMemberCount) {
		t.Fatalf("expected ErrChatPrivateMemberCount, got %v", err)
	}
}

func TestChatServiceCreateGroupRequiresName(t *testing.T) {
	repo := &fakeChatRepo{
		users: map[uint]*userrepo.UserProfile{
			1: {ID: 1, Username: "alice"},
			2: {ID: 2, Username: "bob"},
		},
	}
	svc := chatService{repo: repo}

	_, err := svc.CreateRoom(context.Background(), 1, &chat.CreateChatRoomRequest{
		Type:          chat.ChatRoomType_CHAT_ROOM_TYPE_GROUP,
		MemberUserIds: []string{"2"},
	})
	if !errors.Is(err, ErrChatGroupNameRequired) {
		t.Fatalf("expected ErrChatGroupNameRequired, got %v", err)
	}
}

func TestChatServiceLeavePrivateRoomAllowsCreator(t *testing.T) {
	repo := &fakeChatRepo{
		rooms: map[uint]*model.ChatRoom{
			10: {ID: 10, Type: chatRoomTypePrivate},
		},
		members: map[uint]map[uint]*model.ChatRoomMember{
			10: {
				1: {RoomID: 10, UserID: 1, Role: chatMemberRoleMember},
			},
		},
		deletedMember: true,
	}
	svc := chatService{repo: repo}

	if err := svc.LeaveRoom(context.Background(), 1, 10); err != nil {
		t.Fatalf("LeaveRoom returned error: %v", err)
	}
}

func TestChatServiceMarkRoomReadMapsMissingMessage(t *testing.T) {
	repo := &fakeChatRepo{
		rooms: map[uint]*model.ChatRoom{
			10: {ID: 10, Type: chatRoomTypeGroup},
		},
		members: map[uint]map[uint]*model.ChatRoomMember{
			10: {
				1: {RoomID: 10, UserID: 1, Role: chatMemberRoleMember},
			},
		},
		updateReadErr: gorm.ErrRecordNotFound,
	}
	svc := chatService{repo: repo}

	err := svc.MarkRoomRead(context.Background(), 1, 10, 99)
	if !errors.Is(err, ErrChatMessageNotFound) {
		t.Fatalf("expected ErrChatMessageNotFound, got %v", err)
	}
}

func TestChatServiceInviteMembersRequiresOwnerOrAdmin(t *testing.T) {
	repo := &fakeChatRepo{
		users: map[uint]*userrepo.UserProfile{
			2: {ID: 2, Username: "bob"},
		},
		rooms: map[uint]*model.ChatRoom{
			10: {ID: 10, Type: chatRoomTypeGroup},
		},
		members: map[uint]map[uint]*model.ChatRoomMember{
			10: {
				1: {RoomID: 10, UserID: 1, Role: chatMemberRoleMember},
			},
		},
	}
	svc := chatService{repo: repo}

	err := svc.InviteMembers(context.Background(), 1, 10, []string{"2"})
	if !errors.Is(err, interactionsvc.ErrNoPermission) {
		t.Fatalf("expected ErrNoPermission, got %v", err)
	}
}

func TestChatServiceMarkRoomReadAllowsIdempotentUpdate(t *testing.T) {
	repo := &fakeChatRepo{
		rooms: map[uint]*model.ChatRoom{
			10: {ID: 10, Type: chatRoomTypeGroup},
		},
		members: map[uint]map[uint]*model.ChatRoomMember{
			10: {
				1: {RoomID: 10, UserID: 1, Role: chatMemberRoleMember},
			},
		},
	}
	svc := chatService{repo: repo}

	if err := svc.MarkRoomRead(context.Background(), 1, 10, 99); err != nil {
		t.Fatalf("MarkRoomRead returned error: %v", err)
	}
}

func TestChatServiceCreateMessageRejectsEmptyContent(t *testing.T) {
	repo := &fakeChatRepo{
		rooms: map[uint]*model.ChatRoom{
			10: {ID: 10, Type: chatRoomTypeGroup},
		},
		members: map[uint]map[uint]*model.ChatRoomMember{
			10: {
				1: {RoomID: 10, UserID: 1, Role: chatMemberRoleMember},
			},
		},
	}
	svc := chatService{repo: repo}

	_, err := svc.CreateMessage(context.Background(), 1, 10, "   ", "client-msg-id")
	if !errors.Is(err, interactionsvc.ErrCommentEmpty) {
		t.Fatalf("expected ErrCommentEmpty, got %v", err)
	}
}
