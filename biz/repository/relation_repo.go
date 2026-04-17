package repository

import (
	"context"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/rdb"
)

func FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	if err := db.FollowUser(ctx, fromUserID, toUserID); err != nil {
		return err
	}

	deleteRelationCaches(ctx, fromUserID, toUserID)
	return nil
}

func UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	deleted, err := db.UnfollowUser(ctx, fromUserID, toUserID)
	if err != nil {
		return false, err
	}
	if deleted {
		deleteRelationCaches(ctx, fromUserID, toUserID)
	}
	return deleted, nil
}

func ListFollowings(ctx context.Context, userID uint, offset, limit int) ([]UserProfile, int64, error) {
	version, err := rdb.GetRelationFollowingCacheVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	if cached, ok, err := rdb.GetRelationFollowingCache(ctx, userID, version, offset, limit); err == nil && ok {
		users, err := ListUserSnapshotsByIDs(ctx, cached.UserIDs)
		if err != nil {
			return nil, 0, err
		}
		return users, cached.Total, nil
	}

	userIDs, total, err := db.ListFollowingIDs(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	users, err := ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	_ = rdb.SetRelationFollowingCache(ctx, userID, version, offset, limit, rdb.RelationIDListCache{
		UserIDs: userIDs,
		Total:   total,
	})
	return users, total, nil
}

func ListFollowers(ctx context.Context, userID uint, offset, limit int) ([]UserProfile, int64, error) {
	version, err := rdb.GetRelationFollowerCacheVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	if cached, ok, err := rdb.GetRelationFollowerCache(ctx, userID, version, offset, limit); err == nil && ok {
		users, err := ListUserSnapshotsByIDs(ctx, cached.UserIDs)
		if err != nil {
			return nil, 0, err
		}
		return users, cached.Total, nil
	}

	userIDs, total, err := db.ListFollowerIDs(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	users, err := ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	_ = rdb.SetRelationFollowerCache(ctx, userID, version, offset, limit, rdb.RelationIDListCache{
		UserIDs: userIDs,
		Total:   total,
	})
	return users, total, nil
}

func ListFriends(ctx context.Context, userID uint, offset, limit int) ([]UserProfile, int64, error) {
	version, err := rdb.GetRelationFriendCacheVersion(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	if cached, ok, err := rdb.GetRelationFriendCache(ctx, userID, version, offset, limit); err == nil && ok {
		users, err := ListUserSnapshotsByIDs(ctx, cached.UserIDs)
		if err != nil {
			return nil, 0, err
		}
		return users, cached.Total, nil
	}

	userIDs, total, err := db.ListFriendIDs(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	users, err := ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	_ = rdb.SetRelationFriendCache(ctx, userID, version, offset, limit, rdb.RelationIDListCache{
		UserIDs: userIDs,
		Total:   total,
	})
	return users, total, nil
}

func deleteRelationCaches(ctx context.Context, fromUserID, toUserID uint) {
	_ = rdb.BumpFollowingCacheVersion(ctx, fromUserID)
	_ = rdb.BumpFollowerCacheVersion(ctx, toUserID)
	_ = rdb.BumpFriendCacheVersion(ctx, fromUserID)
	_ = rdb.BumpFriendCacheVersion(ctx, toUserID)
	_ = rdb.DeleteUserProfileCache(ctx, fromUserID)
	_ = rdb.DeleteUserProfileCache(ctx, toUserID)
}
