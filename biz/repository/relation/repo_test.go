package relation

import (
	"context"
	"errors"
	"testing"
	dbdal "video-platform/biz/dal/db"
	rdbdal "video-platform/biz/dal/rdb"
	userrepo "video-platform/biz/repository/user"
)

type fakeRelationDBStore struct {
	followUserFn       func(ctx context.Context, fromUserID, toUserID uint) error
	unfollowUserFn     func(ctx context.Context, fromUserID, toUserID uint) (bool, error)
	listFollowingIDsFn func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error)
	listFollowerIDsFn  func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error)
	listFriendIDsFn    func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error)
}

func (f fakeRelationDBStore) FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	return f.followUserFn(ctx, fromUserID, toUserID)
}

func (f fakeRelationDBStore) UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	return f.unfollowUserFn(ctx, fromUserID, toUserID)
}

func (f fakeRelationDBStore) ListFollowingIDs(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error) {
	return f.listFollowingIDsFn(ctx, userID, cursor, limit)
}

func (f fakeRelationDBStore) ListFollowerIDs(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error) {
	return f.listFollowerIDsFn(ctx, userID, cursor, limit)
}

func (f fakeRelationDBStore) ListFriendIDs(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error) {
	return f.listFriendIDsFn(ctx, userID, cursor, limit)
}

type fakeRelationCacheStore struct {
	getRelationFollowingCacheVersionFn func(ctx context.Context, userID uint) (int64, error)
	getRelationFollowingCacheFn        func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error)
	setRelationFollowingCacheFn        func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error
	getRelationFollowerCacheVersionFn  func(ctx context.Context, userID uint) (int64, error)
	getRelationFollowerCacheFn         func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error)
	setRelationFollowerCacheFn         func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error
	getRelationFriendCacheVersionFn    func(ctx context.Context, userID uint) (int64, error)
	getRelationFriendCacheFn           func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error)
	setRelationFriendCacheFn           func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error
	bumpFollowingCacheVersionFn        func(ctx context.Context, userID uint) error
	bumpFollowerCacheVersionFn         func(ctx context.Context, userID uint) error
	bumpFriendCacheVersionFn           func(ctx context.Context, userID uint) error
	deleteUserProfileCacheFn           func(ctx context.Context, userID uint) error
}

func (f fakeRelationCacheStore) GetRelationFollowingCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return f.getRelationFollowingCacheVersionFn(ctx, userID)
}

func (f fakeRelationCacheStore) GetRelationFollowingCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
	return f.getRelationFollowingCacheFn(ctx, userID, version, cursor, limit)
}

func (f fakeRelationCacheStore) SetRelationFollowingCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
	return f.setRelationFollowingCacheFn(ctx, userID, version, cursor, limit, value)
}

func (f fakeRelationCacheStore) GetRelationFollowerCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return f.getRelationFollowerCacheVersionFn(ctx, userID)
}

func (f fakeRelationCacheStore) GetRelationFollowerCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
	return f.getRelationFollowerCacheFn(ctx, userID, version, cursor, limit)
}

func (f fakeRelationCacheStore) SetRelationFollowerCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
	return f.setRelationFollowerCacheFn(ctx, userID, version, cursor, limit, value)
}

func (f fakeRelationCacheStore) GetRelationFriendCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return f.getRelationFriendCacheVersionFn(ctx, userID)
}

func (f fakeRelationCacheStore) GetRelationFriendCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
	return f.getRelationFriendCacheFn(ctx, userID, version, cursor, limit)
}

func (f fakeRelationCacheStore) SetRelationFriendCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
	return f.setRelationFriendCacheFn(ctx, userID, version, cursor, limit, value)
}

func (f fakeRelationCacheStore) BumpFollowingCacheVersion(ctx context.Context, userID uint) error {
	return f.bumpFollowingCacheVersionFn(ctx, userID)
}

func (f fakeRelationCacheStore) BumpFollowerCacheVersion(ctx context.Context, userID uint) error {
	return f.bumpFollowerCacheVersionFn(ctx, userID)
}

func (f fakeRelationCacheStore) BumpFriendCacheVersion(ctx context.Context, userID uint) error {
	return f.bumpFriendCacheVersionFn(ctx, userID)
}

func (f fakeRelationCacheStore) DeleteUserProfileCache(ctx context.Context, userID uint) error {
	return f.deleteUserProfileCacheFn(ctx, userID)
}

type fakeRelationSnapshotStore struct {
	listUserSnapshotsByIDsFn func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error)
}

func (f fakeRelationSnapshotStore) ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
	return f.listUserSnapshotsByIDsFn(ctx, userIDs)
}

func TestRelationStoreListFollowingsUsesCache(t *testing.T) {
	store := relationStore{
		db: fakeRelationDBStore{
			listFollowingIDsFn: func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error) {
				t.Fatal("db should not be called on cache hit")
				return nil, 0, nil
			},
		},
		cache: fakeRelationCacheStore{
			getRelationFollowingCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) {
				return 2, nil
			},
			getRelationFollowingCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return &rdbdal.RelationIDListCache{
					UserIDs:    []uint{5, 6},
					Total:      8,
					NextCursor: 20,
					HasMore:    true,
				}, true, nil
			},
			setRelationFollowingCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				t.Fatal("cache set should not be called on cache hit")
				return nil
			},
			getRelationFollowerCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) { return 0, nil },
			getRelationFollowerCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return nil, false, nil
			},
			setRelationFollowerCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				return nil
			},
			getRelationFriendCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) { return 0, nil },
			getRelationFriendCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return nil, false, nil
			},
			setRelationFriendCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				return nil
			},
			bumpFollowingCacheVersionFn: func(ctx context.Context, userID uint) error { return nil },
			bumpFollowerCacheVersionFn:  func(ctx context.Context, userID uint) error { return nil },
			bumpFriendCacheVersionFn:    func(ctx context.Context, userID uint) error { return nil },
			deleteUserProfileCacheFn:    func(ctx context.Context, userID uint) error { return nil },
		},
		snapshots: fakeRelationSnapshotStore{
			listUserSnapshotsByIDsFn: func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
				if len(userIDs) != 2 || userIDs[0] != 5 || userIDs[1] != 6 {
					t.Fatalf("unexpected userIDs: %+v", userIDs)
				}
				return []userrepo.UserProfile{
					{ID: 5, Username: "alice"},
					{ID: 6, Username: "bob"},
				}, nil
			},
		},
	}

	got, err := store.ListFollowings(context.Background(), 1, 0, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Users) != 2 || got.Users[0].Username != "alice" {
		t.Fatalf("unexpected users: %+v", got.Users)
	}
	if got.Total != 8 || got.NextCursor != 20 || !got.HasMore {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestRelationStoreListFollowersFallsBackToDBAndSetsCache(t *testing.T) {
	var cachedValue rdbdal.RelationIDListCache

	store := relationStore{
		db: fakeRelationDBStore{
			listFollowerIDsFn: func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error) {
				if limit != 2 {
					t.Fatalf("expected limit %d, got %d", 2, limit)
				}
				return []dbdal.RelationIDItem{
					{RelationID: 11, UserID: 21},
					{RelationID: 10, UserID: 22},
				}, 3, nil
			},
		},
		cache: fakeRelationCacheStore{
			getRelationFollowingCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) { return 0, nil },
			getRelationFollowingCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return nil, false, nil
			},
			setRelationFollowingCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				return nil
			},
			getRelationFollowerCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) {
				return 4, nil
			},
			getRelationFollowerCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return nil, false, nil
			},
			setRelationFollowerCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				cachedValue = value.(rdbdal.RelationIDListCache)
				return nil
			},
			getRelationFriendCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) { return 0, nil },
			getRelationFriendCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return nil, false, nil
			},
			setRelationFriendCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				return nil
			},
			bumpFollowingCacheVersionFn: func(ctx context.Context, userID uint) error { return nil },
			bumpFollowerCacheVersionFn:  func(ctx context.Context, userID uint) error { return nil },
			bumpFriendCacheVersionFn:    func(ctx context.Context, userID uint) error { return nil },
			deleteUserProfileCacheFn:    func(ctx context.Context, userID uint) error { return nil },
		},
		snapshots: fakeRelationSnapshotStore{
			listUserSnapshotsByIDsFn: func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) {
				if len(userIDs) != 1 || userIDs[0] != 21 {
					t.Fatalf("unexpected userIDs: %+v", userIDs)
				}
				return []userrepo.UserProfile{{ID: 21, Username: "fan"}}, nil
			},
		},
	}

	got, err := store.ListFollowers(context.Background(), 2, 0, 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Users) != 1 || got.Users[0].Username != "fan" {
		t.Fatalf("unexpected users: %+v", got.Users)
	}
	if !got.HasMore || got.NextCursor != 11 || got.Total != 3 {
		t.Fatalf("unexpected result: %+v", got)
	}
	if len(cachedValue.UserIDs) != 1 || cachedValue.UserIDs[0] != 21 || cachedValue.NextCursor != 11 {
		t.Fatalf("unexpected cached value: %+v", cachedValue)
	}
}

func TestRelationStoreFollowUserDeletesCaches(t *testing.T) {
	var followingUserID uint
	var followerUserID uint
	friendBumps := make([]uint, 0, 2)
	deletedUsers := make([]uint, 0, 2)

	store := relationStore{
		db: fakeRelationDBStore{
			followUserFn: func(ctx context.Context, fromUserID, toUserID uint) error {
				return nil
			},
		},
		cache: fakeRelationCacheStore{
			getRelationFollowingCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) { return 0, nil },
			getRelationFollowingCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return nil, false, nil
			},
			setRelationFollowingCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				return nil
			},
			getRelationFollowerCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) { return 0, nil },
			getRelationFollowerCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return nil, false, nil
			},
			setRelationFollowerCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				return nil
			},
			getRelationFriendCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) { return 0, nil },
			getRelationFriendCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return nil, false, nil
			},
			setRelationFriendCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				return nil
			},
			bumpFollowingCacheVersionFn: func(ctx context.Context, userID uint) error {
				followingUserID = userID
				return nil
			},
			bumpFollowerCacheVersionFn: func(ctx context.Context, userID uint) error {
				followerUserID = userID
				return nil
			},
			bumpFriendCacheVersionFn: func(ctx context.Context, userID uint) error {
				friendBumps = append(friendBumps, userID)
				return nil
			},
			deleteUserProfileCacheFn: func(ctx context.Context, userID uint) error {
				deletedUsers = append(deletedUsers, userID)
				return nil
			},
		},
		snapshots: fakeRelationSnapshotStore{
			listUserSnapshotsByIDsFn: func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) { return nil, nil },
		},
	}

	if err := store.FollowUser(context.Background(), 3, 9); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if followingUserID != 3 || followerUserID != 9 {
		t.Fatalf("unexpected relation cache bumps following=%d follower=%d", followingUserID, followerUserID)
	}
	if len(friendBumps) != 2 || friendBumps[0] != 3 || friendBumps[1] != 9 {
		t.Fatalf("unexpected friend bumps: %+v", friendBumps)
	}
	if len(deletedUsers) != 2 || deletedUsers[0] != 3 || deletedUsers[1] != 9 {
		t.Fatalf("unexpected deleted user caches: %+v", deletedUsers)
	}
}

func TestRelationStoreListFriendsReturnsVersionError(t *testing.T) {
	wantErr := errors.New("version error")
	store := relationStore{
		db: fakeRelationDBStore{},
		cache: fakeRelationCacheStore{
			getRelationFollowingCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) { return 0, nil },
			getRelationFollowingCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return nil, false, nil
			},
			setRelationFollowingCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				return nil
			},
			getRelationFollowerCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) { return 0, nil },
			getRelationFollowerCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return nil, false, nil
			},
			setRelationFollowerCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				return nil
			},
			getRelationFriendCacheVersionFn: func(ctx context.Context, userID uint) (int64, error) {
				return 0, wantErr
			},
			getRelationFriendCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error) {
				return nil, false, nil
			},
			setRelationFriendCacheFn: func(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
				return nil
			},
			bumpFollowingCacheVersionFn: func(ctx context.Context, userID uint) error { return nil },
			bumpFollowerCacheVersionFn:  func(ctx context.Context, userID uint) error { return nil },
			bumpFriendCacheVersionFn:    func(ctx context.Context, userID uint) error { return nil },
			deleteUserProfileCacheFn:    func(ctx context.Context, userID uint) error { return nil },
		},
		snapshots: fakeRelationSnapshotStore{
			listUserSnapshotsByIDsFn: func(ctx context.Context, userIDs []uint) ([]userrepo.UserProfile, error) { return nil, nil },
		},
	}

	if _, err := store.ListFriends(context.Background(), 1, 0, 10); !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
