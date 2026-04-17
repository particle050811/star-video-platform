package db

import (
	"context"
	"video-platform/biz/dal/model"
	"video-platform/pkg/constant"

	"gorm.io/gorm"
)

type UserSnapshot struct {
	ID        uint
	Username  string
	AvatarURL string
}

func CreateUser(ctx context.Context, user *model.User) error {
	return DB.WithContext(ctx).Create(user).Error
}

func GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := DB.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error) {
	var userIDs []uint
	if err := DB.WithContext(ctx).Model(&model.User{}).
		Where("username LIKE ?", "%"+username+"%").
		Pluck("id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}

func ListUsersByIDs(ctx context.Context, userIDs []uint) ([]model.User, error) {
	users := make([]model.User, 0, len(userIDs))
	if len(userIDs) == 0 {
		return users, nil
	}

	if err := DB.WithContext(ctx).
		Where("id IN ?", userIDs).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]UserSnapshot, error) {
	users, err := ListUsersByIDs(ctx, userIDs)
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

func GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	if err := DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	tx := DB.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("avatar_url", avatarURL)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
