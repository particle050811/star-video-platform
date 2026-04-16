package service

import (
	"context"
	"errors"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	"video-platform/pkg/auth"

	"gorm.io/gorm"
)

// Register 用户注册
func Register(ctx context.Context, username, password string) error {
	hashedPassword, err := auth.HashPassword(password)
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

func Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error) {
	user, err := db.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", ErrUserNotFound
		}
		return "", "", err
	}

	if err := auth.CheckPassword(user.Password, password); err != nil {
		return "", "", ErrPasswordWrong
	}

	return auth.GenerateTokenPair(user.ID)
}
