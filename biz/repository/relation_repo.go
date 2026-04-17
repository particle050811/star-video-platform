package repository

import (
	"context"
	"video-platform/biz/dal/db"
)

func FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	return db.FollowUser(ctx, fromUserID, toUserID)
}

func UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	return db.UnfollowUser(ctx, fromUserID, toUserID)
}

func ListFollowings(ctx context.Context, userID uint, offset, limit int) ([]UserSnapshot, int64, error) {
	userIDs, total, err := db.ListFollowingIDs(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	users, err := ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func ListFollowers(ctx context.Context, userID uint, offset, limit int) ([]UserSnapshot, int64, error) {
	userIDs, total, err := db.ListFollowerIDs(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	users, err := ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func ListFriends(ctx context.Context, userID uint, offset, limit int) ([]UserSnapshot, int64, error) {
	userIDs, total, err := db.ListFriendIDs(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	users, err := ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
