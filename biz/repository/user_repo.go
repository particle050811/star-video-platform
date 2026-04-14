package repository

import (
	"context"
	"video-platform/biz/dal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
}
