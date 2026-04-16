package service

import (
	"context"
	"errors"
	"strconv"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	v1 "video-platform/biz/model/platform"
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

func RefreshToken(ctx context.Context, refresh_token string) (accessToken, refreshToken string, err error) {
	accessToken, refreshToken, err = auth.RefreshTokens(refresh_token)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			return "", "", ErrTokenExpired
		}
		if errors.Is(err, auth.ErrTokenInvalid) || errors.Is(err, auth.ErrTokenMalformed) {
			return "", "", ErrTokenInvalid
		}
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func GetUserInfo(ctx context.Context, userID uint) (*v1.User, error) {
	user, err := db.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &v1.User{
		Id:        strconv.FormatUint(uint64(user.ID), 10),
		Username:  user.Username,
		AvatarUrl: user.AvatarURL,
	}, nil
}
