package db

import (
	"context"
	"video-platform/biz/dal/model"
)

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

func GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	if err := DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
