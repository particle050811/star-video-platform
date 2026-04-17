package service

import (
	"context"
	"errors"
	"mime/multipart"
	"strconv"
	"video-platform/biz/dal/model"
	v1 "video-platform/biz/model/user"
	"video-platform/pkg/auth"
	"video-platform/pkg/upload"

	"gorm.io/gorm"
)

type userService struct {
	repo   userRepository
	auth   authProvider
	upload uploadProvider
}

var defaultUserService = userService{
	repo:   defaultUserRepository{},
	auth:   defaultAuthProvider{},
	upload: defaultUploadProvider{},
}

var User = defaultUserService

func (s userService) Register(ctx context.Context, username, password string) error {
	hashedPassword, err := s.auth.HashPassword(password)
	if err != nil {
		return err
	}

	if err := s.repo.CreateUser(ctx, &model.User{
		Username: username, Password: string(hashedPassword)}); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrUserExists
		}
		return err
	}
	return nil
}

func (s userService) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", ErrUserNotFound
		}
		return "", "", err
	}

	if err := s.auth.CheckPassword(user.Password, password); err != nil {
		return "", "", ErrPasswordWrong
	}

	return s.auth.GenerateTokenPair(user.ID)
}

func (s userService) RefreshToken(ctx context.Context, refreshToken string) (accessToken, nextRefreshToken string, err error) {
	accessToken, nextRefreshToken, err = s.auth.RefreshTokens(refreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			return "", "", ErrTokenExpired
		}
		if errors.Is(err, auth.ErrTokenInvalid) || errors.Is(err, auth.ErrTokenMalformed) {
			return "", "", ErrTokenInvalid
		}
		return "", "", err
	}

	return accessToken, nextRefreshToken, nil
}

func (s userService) GetUserInfo(ctx context.Context, userID uint) (*v1.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
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

func (s userService) UpdateUserAvatar(ctx context.Context, userID uint, file *multipart.FileHeader) (err error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	var avatarURL string
	defer func() {
		if err != nil {
			_ = s.upload.RemoveAvatar(avatarURL)
		}
	}()

	var savePath string
	savePath, avatarURL, err = s.upload.PrepareAvatar(userID, file.Filename)
	if err != nil {
		if errors.Is(err, upload.ErrUnsupportedAvatarExt) {
			return ErrUnsupportedAvatarExt
		}
		return err
	}

	if err := s.upload.SaveFile(file, savePath); err != nil {
		return err
	}

	if err := s.repo.UpdateUserAvatar(ctx, userID, avatarURL); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if user.AvatarURL != avatarURL {
		_ = s.upload.RemoveAvatar(user.AvatarURL)
	}

	return nil
}
