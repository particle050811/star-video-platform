package service

import (
	"context"
	"errors"
	"testing"
	"video-platform/biz/repository"

	"gorm.io/gorm"
)

type fakeRelationRepository struct {
	getUserByIDFn    func(ctx context.Context, userID uint) (*repository.UserProfile, error)
	followUserFn     func(ctx context.Context, fromUserID, toUserID uint) error
	unfollowUserFn   func(ctx context.Context, fromUserID, toUserID uint) (bool, error)
	listFollowingsFn func(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error)
	listFollowersFn  func(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error)
	listFriendsFn    func(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error)
}

func (f fakeRelationRepository) GetUserByID(ctx context.Context, userID uint) (*repository.UserProfile, error) {
	return f.getUserByIDFn(ctx, userID)
}

func (f fakeRelationRepository) FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	return f.followUserFn(ctx, fromUserID, toUserID)
}

func (f fakeRelationRepository) UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	return f.unfollowUserFn(ctx, fromUserID, toUserID)
}

func (f fakeRelationRepository) ListFollowings(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error) {
	return f.listFollowingsFn(ctx, userID, offset, limit)
}

func (f fakeRelationRepository) ListFollowers(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error) {
	return f.listFollowersFn(ctx, userID, offset, limit)
}

func (f fakeRelationRepository) ListFriends(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error) {
	return f.listFriendsFn(ctx, userID, offset, limit)
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
			getUserByIDFn: func(ctx context.Context, userID uint) (*repository.UserProfile, error) {
				return &repository.UserProfile{ID: userID}, nil
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
			getUserByIDFn: func(ctx context.Context, userID uint) (*repository.UserProfile, error) {
				return &repository.UserProfile{ID: userID}, nil
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

func TestRelationServiceListFollowingsUsesPaginationAndBuildsResponse(t *testing.T) {
	svc := relationService{
		repo: fakeRelationRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*repository.UserProfile, error) {
				return &repository.UserProfile{ID: userID}, nil
			},
			listFollowingsFn: func(ctx context.Context, userID uint, offset, limit int) ([]repository.UserProfile, int64, error) {
				if userID != 5 {
					t.Fatalf("expected user id %d, got %d", 5, userID)
				}
				if offset != 20 {
					t.Fatalf("expected offset %d, got %d", 20, offset)
				}
				if limit != 20 {
					t.Fatalf("expected limit %d, got %d", 20, limit)
				}
				return []repository.UserProfile{
					{
						ID:        9,
						Username:  "alice",
						AvatarURL: "/static/avatars/a.png",
					},
				}, 1, nil
			},
		},
	}

	got, err := svc.ListFollowings(context.Background(), 5, 2, 0)
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
}
