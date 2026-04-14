package db

import (
	"context"
	"video-platform/biz/dal/model"
)

func CreateUser(ctx context.Context, user *model.User) error {
	return DB.WithContext(ctx).Create(user).Error

}
