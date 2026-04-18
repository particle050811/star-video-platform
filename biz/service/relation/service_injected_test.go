package relation

import (
	"context"
	"errors"
	"testing"
	relationrepo "video-platform/biz/repository/relation"
	userrepo "video-platform/biz/repository/user"

	"gorm.io/gorm"
)

type fakeRelationRepository struct {
	getUserByIDFn    func(ctx context.Context, userID uint) (*userrepo.UserProfile, error)
	followUserFn     func(ctx context.Context, fromUserID, toUserID uint) error
	unfollowUserFn   func(ctx context.Context, fromUserID, toUserID uint) (bool, error)
	listFollowingsFn func(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error)
	listFollowersFn  func(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error)
	listFriendsFn    func(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error)
}

func (f fakeRelationRepository) GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
	return f.getUserByIDFn(ctx, userID)
}

func (f fakeRelationRepository) FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	return f.followUserFn(ctx, fromUserID, toUserID)
}

func (f fakeRelationRepository) UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	return f.unfollowUserFn(ctx, fromUserID, toUserID)
}

func (f fakeRelationRepository) ListFollowings(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error) {
	return f.listFollowingsFn(ctx, userID, cursor, limit)
}

func (f fakeRelationRepository) ListFollowers(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error) {
	return f.listFollowersFn(ctx, userID, cursor, limit)
}

func (f fakeRelationRepository) ListFriends(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error) {
	return f.listFriendsFn(ctx, userID, cursor, limit)
}

func TestRelationServiceRelationActionRejectsSelfFollow(t *testing.T) {
	svc := relationService{}

	err := svc.RelationAction(context.Background(), 1, 1, relationActionFollow)
	if !errors.Is(err, ErrCannotFollowSelf) {
		t.Fatalf("expected error %v, got %v", ErrCannotFollowSelf, err)
	}
}

func TestRelationServiceRelationActionMapsFollowDuplicate(t *testing.T) {
	svc := relationService{
		repo: fakeRelationRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
				return &userrepo.UserProfile{ID: userID}, nil
			},
			followUserFn: func(ctx context.Context, fromUserID, toUserID uint) error {
				if fromUserID != 1 || toUserID != 2 {
					t.Fatalf("unexpected follow params from=%d to=%d", fromUserID, toUserID)
				}
				return gorm.ErrDuplicatedKey
			},
		},
	}

	err := svc.RelationAction(context.Background(), 1, 2, relationActionFollow)
	if !errors.Is(err, ErrAlreadyFollowed) {
		t.Fatalf("expected error %v, got %v", ErrAlreadyFollowed, err)
	}
}

func TestRelationServiceRelationActionMapsUnfollowMiss(t *testing.T) {
	svc := relationService{
		repo: fakeRelationRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
				return &userrepo.UserProfile{ID: userID}, nil
			},
			unfollowUserFn: func(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
				return false, nil
			},
		},
	}

	err := svc.RelationAction(context.Background(), 1, 2, relationActionUnfollow)
	if !errors.Is(err, ErrFollowNotFound) {
		t.Fatalf("expected error %v, got %v", ErrFollowNotFound, err)
	}
}

func TestRelationServiceListFollowingsUsesCursorAndBuildsResponse(t *testing.T) {
	svc := relationService{
		repo: fakeRelationRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
				return &userrepo.UserProfile{ID: userID}, nil
			},
			listFollowingsFn: func(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error) {
				if userID != 5 {
					t.Fatalf("expected user id %d, got %d", 5, userID)
				}
				if cursor != 12 {
					t.Fatalf("expected cursor %d, got %d", 12, cursor)
				}
				if limit != 20 {
					t.Fatalf("expected limit %d, got %d", 20, limit)
				}
				return &relationrepo.RelationListResult{
					Users: []userrepo.UserProfile{{
						ID:        9,
						Username:  "alice",
						AvatarURL: "/static/avatars/a.png",
					}},
					Total:      1,
					NextCursor: 9,
					HasMore:    true,
				}, nil
			},
		},
	}

	got, err := svc.ListFollowings(context.Background(), 5, 12, 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Total != 1 {
		t.Fatalf("expected total %d, got %d", 1, got.Total)
	}
	if len(got.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got.Items))
	}
	if got.Items[0].Id != "9" {
		t.Fatalf("expected item id %q, got %q", "9", got.Items[0].Id)
	}
	if got.Items[0].Username != "alice" {
		t.Fatalf("expected username %q, got %q", "alice", got.Items[0].Username)
	}
	if got.NextCursor != "9" || !got.HasMore {
		t.Fatalf("unexpected cursor response: %+v", got)
	}
}
