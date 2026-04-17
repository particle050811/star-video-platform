package repository

import (
	"context"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	"video-platform/biz/dal/rdb"
	"video-platform/pkg/constant"
)

type UserProfile struct {
	ID             uint
	Username       string
	AvatarURL      string
	FollowingCount int64
	FollowerCount  int64
}

func ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error) {
	return db.ListUserIDsByUsername(ctx, username)
}

func CreateUser(ctx context.Context, user *model.User) error {
	return db.CreateUser(ctx, user)
}

func GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return db.GetUserByUsername(ctx, username)
}

func GetUserByID(ctx context.Context, userID uint) (*UserProfile, error) {
	if cached, ok, err := rdb.GetUserProfileCache(ctx, userID); err == nil && ok {
		return cachedUserToProfile(cached), nil
	}

	fetched, err := db.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	_ = rdb.SetUserProfileCache(ctx, userID, newUserCachePayload(fetched))
	return newUserProfile(fetched), nil
}

func UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	if err := db.UpdateUserAvatar(ctx, userID, avatarURL); err != nil {
		return err
	}

	_ = rdb.DeleteUserProfileCache(ctx, userID)
	return nil
}

func ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]UserProfile, error) {
	userMap := make(map[uint]UserProfile, len(userIDs))
	missedUserIDs := make([]uint, 0, len(userIDs))

	for _, userID := range userIDs {
		if user, ok, err := rdb.GetUserProfileCache(ctx, userID); err == nil && ok {
			userMap[userID] = UserProfile{
				ID:             user.ID,
				Username:       user.Username,
				AvatarURL:      user.AvatarURL,
				FollowingCount: user.FollowingCount,
				FollowerCount:  user.FollowerCount,
			}
			continue
		}

		missedUserIDs = append(missedUserIDs, userID)
	}

	if len(missedUserIDs) > 0 {
		users, err := db.ListUsersByIDs(ctx, missedUserIDs)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			userMap[user.ID] = UserProfile{
				ID:             user.ID,
				Username:       user.Username,
				AvatarURL:      user.AvatarURL,
				FollowingCount: user.FollowingCount,
				FollowerCount:  user.FollowerCount,
			}
			_ = rdb.SetUserProfileCache(ctx, user.ID, newUserCachePayload(&user))
		}
	}

	snapshots := make([]UserProfile, 0, len(userIDs))
	for _, userID := range userIDs {
		user, ok := userMap[userID]
		if !ok {
			user = UserProfile{
				ID:        userID,
				Username:  constant.DeletedUserName,
				AvatarURL: "",
			}
		}
		snapshots = append(snapshots, user)
	}

	return snapshots, nil
}

func newUserCachePayload(user *model.User) rdb.UserProfileCache {
	return rdb.UserProfileCache{
		ID:             user.ID,
		Username:       user.Username,
		AvatarURL:      user.AvatarURL,
		FollowingCount: user.FollowingCount,
		FollowerCount:  user.FollowerCount,
	}
}

func newUserProfile(user *model.User) *UserProfile {
	return &UserProfile{
		ID:             user.ID,
		Username:       user.Username,
		AvatarURL:      user.AvatarURL,
		FollowingCount: user.FollowingCount,
		FollowerCount:  user.FollowerCount,
	}
}

func cachedUserToProfile(u *rdb.UserProfileCache) *UserProfile {
	return &UserProfile{
		ID:             u.ID,
		Username:       u.Username,
		AvatarURL:      u.AvatarURL,
		FollowingCount: u.FollowingCount,
		FollowerCount:  u.FollowerCount,
	}
}
