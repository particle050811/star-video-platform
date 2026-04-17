package repository

import (
	"context"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	"video-platform/pkg/constant"
)

type UserSnapshot struct {
	ID        uint
	Username  string
	AvatarURL string
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

func GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	return db.GetUserByID(ctx, userID)
}

func UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	return db.UpdateUserAvatar(ctx, userID, avatarURL)
}

func ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]UserSnapshot, error) {
	users, err := db.ListUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	userMap := make(map[uint]UserSnapshot, len(users))
	for _, user := range users {
		userMap[user.ID] = UserSnapshot{
			ID:        user.ID,
			Username:  user.Username,
			AvatarURL: user.AvatarURL,
		}
	}

	snapshots := make([]UserSnapshot, 0, len(userIDs))
	for _, userID := range userIDs {
		user, ok := userMap[userID]
		if !ok {
			user = UserSnapshot{
				ID:        userID,
				Username:  constant.DeletedUserName,
				AvatarURL: "",
			}
		}
		snapshots = append(snapshots, user)
	}

	return snapshots, nil
}
