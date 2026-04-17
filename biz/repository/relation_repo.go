package repository

import (
	"context"
	dbdal "video-platform/biz/dal/db"
	rdbdal "video-platform/biz/dal/rdb"
)

type relationDBStore interface {
	FollowUser(ctx context.Context, fromUserID, toUserID uint) error
	UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error)
	ListFollowingIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error)
	ListFollowerIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error)
	ListFriendIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error)
}

type relationCacheStore interface {
	GetRelationFollowingCacheVersion(ctx context.Context, userID uint) (int64, error)
	GetRelationFollowingCache(ctx context.Context, userID uint, version int64, offset, limit int) (*rdbdal.RelationIDListCache, bool, error)
	SetRelationFollowingCache(ctx context.Context, userID uint, version int64, offset, limit int, value any) error
	GetRelationFollowerCacheVersion(ctx context.Context, userID uint) (int64, error)
	GetRelationFollowerCache(ctx context.Context, userID uint, version int64, offset, limit int) (*rdbdal.RelationIDListCache, bool, error)
	SetRelationFollowerCache(ctx context.Context, userID uint, version int64, offset, limit int, value any) error
	GetRelationFriendCacheVersion(ctx context.Context, userID uint) (int64, error)
	GetRelationFriendCache(ctx context.Context, userID uint, version int64, offset, limit int) (*rdbdal.RelationIDListCache, bool, error)
	SetRelationFriendCache(ctx context.Context, userID uint, version int64, offset, limit int, value any) error
	BumpFollowingCacheVersion(ctx context.Context, userID uint) error
	BumpFollowerCacheVersion(ctx context.Context, userID uint) error
	BumpFriendCacheVersion(ctx context.Context, userID uint) error
	DeleteUserProfileCache(ctx context.Context, userID uint) error
}

type relationSnapshotStore interface {
	ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]UserProfile, error)
}

type relationStore struct {
	db        relationDBStore
	cache     relationCacheStore
	snapshots relationSnapshotStore
}

type defaultRelationSnapshotStore struct{}

func (defaultRelationSnapshotStore) ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]UserProfile, error) {
	return ListUserSnapshotsByIDs(ctx, userIDs)
}

var relations = relationStore{
	db:        dbdal.Relations,
	cache:     rdbdal.Relations,
	snapshots: defaultRelationSnapshotStore{},
}

func FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	return relations.FollowUser(ctx, fromUserID, toUserID)
}

func UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	return relations.UnfollowUser(ctx, fromUserID, toUserID)
}

func ListFollowings(ctx context.Context, userID uint, offset, limit int) ([]UserProfile, int64, error) {
	return relations.ListFollowings(ctx, userID, offset, limit)
}

func ListFollowers(ctx context.Context, userID uint, offset, limit int) ([]UserProfile, int64, error) {
	return relations.ListFollowers(ctx, userID, offset, limit)
}

func ListFriends(ctx context.Context, userID uint, offset, limit int) ([]UserProfile, int64, error) {
	return relations.ListFriends(ctx, userID, offset, limit)
}

func (s relationStore) FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	if err := s.db.FollowUser(ctx, fromUserID, toUserID); err != nil {
		return err
	}

	s.deleteRelationCaches(ctx, fromUserID, toUserID)
	return nil
}

func (s relationStore) UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	deleted, err := s.db.UnfollowUser(ctx, fromUserID, toUserID)
	if err != nil {
		return false, err
	}
	if deleted {
		s.deleteRelationCaches(ctx, fromUserID, toUserID)
	}
	return deleted, nil
}

func (s relationStore) ListFollowings(ctx context.Context, userID uint, offset, limit int) ([]UserProfile, int64, error) {
	version, err := s.cache.GetRelationFollowingCacheVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	if cached, ok, err := s.cache.GetRelationFollowingCache(ctx, userID, version, offset, limit); err == nil && ok {
		users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, cached.UserIDs)
		if err != nil {
			return nil, 0, err
		}
		return users, cached.Total, nil
	}

	userIDs, total, err := s.db.ListFollowingIDs(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	_ = s.cache.SetRelationFollowingCache(ctx, userID, version, offset, limit, rdbdal.RelationIDListCache{
		UserIDs: userIDs,
		Total:   total,
	})
	return users, total, nil
}

func (s relationStore) ListFollowers(ctx context.Context, userID uint, offset, limit int) ([]UserProfile, int64, error) {
	version, err := s.cache.GetRelationFollowerCacheVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	if cached, ok, err := s.cache.GetRelationFollowerCache(ctx, userID, version, offset, limit); err == nil && ok {
		users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, cached.UserIDs)
		if err != nil {
			return nil, 0, err
		}
		return users, cached.Total, nil
	}

	userIDs, total, err := s.db.ListFollowerIDs(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	_ = s.cache.SetRelationFollowerCache(ctx, userID, version, offset, limit, rdbdal.RelationIDListCache{
		UserIDs: userIDs,
		Total:   total,
	})
	return users, total, nil
}

func (s relationStore) ListFriends(ctx context.Context, userID uint, offset, limit int) ([]UserProfile, int64, error) {
	version, err := s.cache.GetRelationFriendCacheVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	if cached, ok, err := s.cache.GetRelationFriendCache(ctx, userID, version, offset, limit); err == nil && ok {
		users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, cached.UserIDs)
		if err != nil {
			return nil, 0, err
		}
		return users, cached.Total, nil
	}

	userIDs, total, err := s.db.ListFriendIDs(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	_ = s.cache.SetRelationFriendCache(ctx, userID, version, offset, limit, rdbdal.RelationIDListCache{
		UserIDs: userIDs,
		Total:   total,
	})
	return users, total, nil
}

func (s relationStore) deleteRelationCaches(ctx context.Context, fromUserID, toUserID uint) {
	_ = s.cache.BumpFollowingCacheVersion(ctx, fromUserID)
	_ = s.cache.BumpFollowerCacheVersion(ctx, toUserID)
	_ = s.cache.BumpFriendCacheVersion(ctx, fromUserID)
	_ = s.cache.BumpFriendCacheVersion(ctx, toUserID)
	_ = s.cache.DeleteUserProfileCache(ctx, fromUserID)
	_ = s.cache.DeleteUserProfileCache(ctx, toUserID)
}
