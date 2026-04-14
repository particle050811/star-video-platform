package service

import (
	"context"
	"errors"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Register 用户注册
func Register(ctx context.Context, username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := db.CreateUser(ctx, &model.User{
		Username: username, Password: string(hashedPassword)}); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrUserExists
		}
		return err
	}
	return nil
}
