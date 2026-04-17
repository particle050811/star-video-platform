package repository

import (
	"context"
	dbdal "video-platform/biz/dal/db"
	rdbdal "video-platform/biz/dal/rdb"
)

type relationDBStore interface {
	FollowUser(ctx context.Context, fromUserID, toUserID uint) error
	UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error)
	ListFollowingIDs(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error)
	ListFollowerIDs(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error)
	ListFriendIDs(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error)
}

type relationCacheStore interface {
	GetRelationFollowingCacheVersion(ctx context.Context, userID uint) (int64, error)
	GetRelationFollowingCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error)
	SetRelationFollowingCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error
	GetRelationFollowerCacheVersion(ctx context.Context, userID uint) (int64, error)
	GetRelationFollowerCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error)
	SetRelationFollowerCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error
	GetRelationFriendCacheVersion(ctx context.Context, userID uint) (int64, error)
	GetRelationFriendCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*rdbdal.RelationIDListCache, bool, error)
	SetRelationFriendCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error
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

type RelationListResult struct {
	Users      []UserProfile
	Total      int64
	NextCursor uint
	HasMore    bool
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

func ListFollowings(ctx context.Context, userID uint, cursor uint, limit int) (*RelationListResult, error) {
	return relations.ListFollowings(ctx, userID, cursor, limit)
}

func ListFollowers(ctx context.Context, userID uint, cursor uint, limit int) (*RelationListResult, error) {
	return relations.ListFollowers(ctx, userID, cursor, limit)
}

func ListFriends(ctx context.Context, userID uint, cursor uint, limit int) (*RelationListResult, error) {
	return relations.ListFriends(ctx, userID, cursor, limit)
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

func (s relationStore) ListFollowings(ctx context.Context, userID uint, cursor uint, limit int) (*RelationListResult, error) {
	version, err := s.cache.GetRelationFollowingCacheVersion(ctx, userID)
	if err != nil {
		return nil, err
	}

	if cached, ok, err := s.cache.GetRelationFollowingCache(ctx, userID, version, cursor, limit); err == nil && ok {
		users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, cached.UserIDs)
		if err != nil {
			return nil, err
		}
		return &RelationListResult{Users: users, Total: cached.Total, NextCursor: cached.NextCursor, HasMore: cached.HasMore}, nil
	}

	userIDs, total, nextCursor, hasMore, err := s.listRelationIDs(ctx, userID, cursor, limit, s.db.ListFollowingIDs)
	if err != nil {
		return nil, err
	}

	users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetRelationFollowingCache(ctx, userID, version, cursor, limit, rdbdal.RelationIDListCache{
		UserIDs:    userIDs,
		Total:      total,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
	return &RelationListResult{Users: users, Total: total, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (s relationStore) ListFollowers(ctx context.Context, userID uint, cursor uint, limit int) (*RelationListResult, error) {
	version, err := s.cache.GetRelationFollowerCacheVersion(ctx, userID)
	if err != nil {
		return nil, err
	}

	if cached, ok, err := s.cache.GetRelationFollowerCache(ctx, userID, version, cursor, limit); err == nil && ok {
		users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, cached.UserIDs)
		if err != nil {
			return nil, err
		}
		return &RelationListResult{Users: users, Total: cached.Total, NextCursor: cached.NextCursor, HasMore: cached.HasMore}, nil
	}

	userIDs, total, nextCursor, hasMore, err := s.listRelationIDs(ctx, userID, cursor, limit, s.db.ListFollowerIDs)
	if err != nil {
		return nil, err
	}

	users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetRelationFollowerCache(ctx, userID, version, cursor, limit, rdbdal.RelationIDListCache{
		UserIDs:    userIDs,
		Total:      total,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
	return &RelationListResult{Users: users, Total: total, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (s relationStore) ListFriends(ctx context.Context, userID uint, cursor uint, limit int) (*RelationListResult, error) {
	version, err := s.cache.GetRelationFriendCacheVersion(ctx, userID)
	if err != nil {
		return nil, err
	}

	if cached, ok, err := s.cache.GetRelationFriendCache(ctx, userID, version, cursor, limit); err == nil && ok {
		users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, cached.UserIDs)
		if err != nil {
			return nil, err
		}
		return &RelationListResult{Users: users, Total: cached.Total, NextCursor: cached.NextCursor, HasMore: cached.HasMore}, nil
	}

	userIDs, total, nextCursor, hasMore, err := s.listRelationIDs(ctx, userID, cursor, limit, s.db.ListFriendIDs)
	if err != nil {
		return nil, err
	}

	users, err := s.snapshots.ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetRelationFriendCache(ctx, userID, version, cursor, limit, rdbdal.RelationIDListCache{
		UserIDs:    userIDs,
		Total:      total,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
	return &RelationListResult{Users: users, Total: total, NextCursor: nextCursor, HasMore: hasMore}, nil
}

type listRelationIDsFunc func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.RelationIDItem, int64, error)

func (s relationStore) listRelationIDs(ctx context.Context, userID uint, cursor uint, limit int, listFn listRelationIDsFunc) ([]uint, int64, uint, bool, error) {
	items, total, err := listFn(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, 0, 0, false, err
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	nextCursor := uint(0)
	if hasMore && len(items) > 0 {
		nextCursor = items[len(items)-1].RelationID
	}

	userIDs := make([]uint, 0, len(items))
	for _, item := range items {
		userIDs = append(userIDs, item.UserID)
	}

	return userIDs, total, nextCursor, hasMore, nil
}

func (s relationStore) deleteRelationCaches(ctx context.Context, fromUserID, toUserID uint) {
	_ = s.cache.BumpFollowingCacheVersion(ctx, fromUserID)
	_ = s.cache.BumpFollowerCacheVersion(ctx, toUserID)
	_ = s.cache.BumpFriendCacheVersion(ctx, fromUserID)
	_ = s.cache.BumpFriendCacheVersion(ctx, toUserID)
	_ = s.cache.DeleteUserProfileCache(ctx, fromUserID)
	_ = s.cache.DeleteUserProfileCache(ctx, toUserID)
}
