package user

import (
	"context"
	"errors"
	"testing"
	"video-platform/biz/dal/model"
	"video-platform/biz/dal/rdb"
	"video-platform/pkg/constant"
)

type fakeUserDBStore struct {
	listUserIDsByUsernameFn func(ctx context.Context, username string) ([]uint, error)
	createUserFn            func(ctx context.Context, user *model.User) error
	getUserByUsernameFn     func(ctx context.Context, username string) (*model.User, error)
	getUserByIDFn           func(ctx context.Context, userID uint) (*model.User, error)
	updateUserAvatarFn      func(ctx context.Context, userID uint, avatarURL string) error
	listUsersByIDsFn        func(ctx context.Context, userIDs []uint) ([]model.User, error)
}

func (f fakeUserDBStore) ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error) {
	return f.listUserIDsByUsernameFn(ctx, username)
}

func (f fakeUserDBStore) CreateUser(ctx context.Context, user *model.User) error {
	return f.createUserFn(ctx, user)
}

func (f fakeUserDBStore) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return f.getUserByUsernameFn(ctx, username)
}

func (f fakeUserDBStore) GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	return f.getUserByIDFn(ctx, userID)
}

func (f fakeUserDBStore) UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	return f.updateUserAvatarFn(ctx, userID, avatarURL)
}

func (f fakeUserDBStore) ListUsersByIDs(ctx context.Context, userIDs []uint) ([]model.User, error) {
	return f.listUsersByIDsFn(ctx, userIDs)
}

type fakeUserCacheStore struct {
	getUserProfileCacheFn    func(ctx context.Context, userID uint) (*rdb.UserProfileCache, bool, error)
	setUserProfileCacheFn    func(ctx context.Context, userID uint, value any) error
	deleteUserProfileCacheFn func(ctx context.Context, userID uint) error
}

func (f fakeUserCacheStore) GetUserProfileCache(ctx context.Context, userID uint) (*rdb.UserProfileCache, bool, error) {
	return f.getUserProfileCacheFn(ctx, userID)
}

func (f fakeUserCacheStore) SetUserProfileCache(ctx context.Context, userID uint, value any) error {
	return f.setUserProfileCacheFn(ctx, userID, value)
}

func (f fakeUserCacheStore) DeleteUserProfileCache(ctx context.Context, userID uint) error {
	return f.deleteUserProfileCacheFn(ctx, userID)
}

func TestUserStoreGetUserByIDUsesCache(t *testing.T) {
	store := userStore{
		db: fakeUserDBStore{
			getUserByIDFn: func(ctx context.Context, userID uint) (*model.User, error) {
				t.Fatal("db should not be called on cache hit")
				return nil, nil
			},
		},
		cache: fakeUserCacheStore{
			getUserProfileCacheFn: func(ctx context.Context, userID uint) (*rdb.UserProfileCache, bool, error) {
				return &rdb.UserProfileCache{
					ID:             userID,
					Username:       "alice",
					AvatarURL:      "/static/avatars/a.png",
					FollowingCount: 3,
					FollowerCount:  4,
				}, true, nil
			},
			setUserProfileCacheFn: func(ctx context.Context, userID uint, value any) error {
				t.Fatal("cache set should not be called on cache hit")
				return nil
			},
		},
	}

	got, err := store.GetUserByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Username != "alice" {
		t.Fatalf("expected username %q, got %q", "alice", got.Username)
	}
}

func TestUserStoreGetUserByIDFallsBackToDBAndSetsCache(t *testing.T) {
	var cachedPayload rdb.UserProfileCache

	store := userStore{
		db: fakeUserDBStore{
			getUserByIDFn: func(ctx context.Context, userID uint) (*model.User, error) {
				return &model.User{
					ID:             userID,
					Username:       "bob",
					AvatarURL:      "/static/avatars/b.png",
					FollowingCount: 5,
					FollowerCount:  6,
				}, nil
			},
		},
		cache: fakeUserCacheStore{
			getUserProfileCacheFn: func(ctx context.Context, userID uint) (*rdb.UserProfileCache, bool, error) {
				return nil, false, nil
			},
			setUserProfileCacheFn: func(ctx context.Context, userID uint, value any) error {
				payload, ok := value.(rdb.UserProfileCache)
				if !ok {
					t.Fatalf("expected cache payload type %T, got %T", rdb.UserProfileCache{}, value)
				}
				cachedPayload = payload
				return nil
			},
		},
	}

	got, err := store.GetUserByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Username != "bob" {
		t.Fatalf("expected username %q, got %q", "bob", got.Username)
	}
	if cachedPayload.ID != 2 || cachedPayload.Username != "bob" {
		t.Fatalf("unexpected cached payload: %+v", cachedPayload)
	}
}

func TestUserStoreUpdateUserAvatarDeletesCache(t *testing.T) {
	var deletedUserID uint

	store := userStore{
		db: fakeUserDBStore{
			updateUserAvatarFn: func(ctx context.Context, userID uint, avatarURL string) error {
				if avatarURL != "/static/avatars/new.png" {
					t.Fatalf("expected avatar url %q, got %q", "/static/avatars/new.png", avatarURL)
				}
				return nil
			},
		},
		cache: fakeUserCacheStore{
			deleteUserProfileCacheFn: func(ctx context.Context, userID uint) error {
				deletedUserID = userID
				return nil
			},
		},
	}

	if err := store.UpdateUserAvatar(context.Background(), 3, "/static/avatars/new.png"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if deletedUserID != 3 {
		t.Fatalf("expected deleted cache user id %d, got %d", 3, deletedUserID)
	}
}

func TestUserStoreListUserSnapshotsByIDsFillsDeletedUsers(t *testing.T) {
	store := userStore{
		db: fakeUserDBStore{
			listUsersByIDsFn: func(ctx context.Context, userIDs []uint) ([]model.User, error) {
				return []model.User{
					{ID: 2, Username: "bob", AvatarURL: "/static/avatars/b.png"},
				}, nil
			},
		},
		cache: fakeUserCacheStore{
			getUserProfileCacheFn: func(ctx context.Context, userID uint) (*rdb.UserProfileCache, bool, error) {
				return nil, false, nil
			},
			setUserProfileCacheFn: func(ctx context.Context, userID uint, value any) error {
				return nil
			},
		},
	}

	got, err := store.ListUserSnapshotsByIDs(context.Background(), []uint{2, 9})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[0].Username != "bob" {
		t.Fatalf("expected first username %q, got %q", "bob", got[0].Username)
	}
	if got[1].Username != constant.DeletedUserName {
		t.Fatalf("expected second username %q, got %q", constant.DeletedUserName, got[1].Username)
	}
}

func TestUserStoreGetUserByIDReturnsDBError(t *testing.T) {
	wantErr := errors.New("db error")
	store := userStore{
		db: fakeUserDBStore{
			getUserByIDFn: func(ctx context.Context, userID uint) (*model.User, error) {
				return nil, wantErr
			},
		},
		cache: fakeUserCacheStore{
			getUserProfileCacheFn: func(ctx context.Context, userID uint) (*rdb.UserProfileCache, bool, error) {
				return nil, false, nil
			},
		},
	}

	_, err := store.GetUserByID(context.Background(), 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
