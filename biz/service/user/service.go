package user

import (
	"context"
	"errors"
	"mime/multipart"
	"strconv"
	"video-platform/biz/dal/model"
	v1 "video-platform/biz/model/user"
	userrepo "video-platform/biz/repository/user"
	"video-platform/pkg/auth"
	"video-platform/pkg/upload"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error)
	UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error
}

type authProvider interface {
	HashPassword(password string) (string, error)
	CheckPassword(hashedPassword, password string) error
	GenerateTokenPair(userID uint) (accessToken, refreshToken string, err error)
	RefreshTokens(refreshToken string) (newAccessToken, newRefreshToken string, err error)
}

type uploadProvider interface {
	PrepareAvatar(userID uint, originalFilename string) (savePath, avatarURL string, err error)
	SaveFile(file *multipart.FileHeader, savePath string) error
	RemoveAvatar(avatarURL string) error
}

type defaultUserRepository struct{}

func (defaultUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	return userrepo.CreateUser(ctx, user)
}

func (defaultUserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return userrepo.GetUserByUsername(ctx, username)
}

func (defaultUserRepository) GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
	return userrepo.GetUserByID(ctx, userID)
}

func (defaultUserRepository) UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	return userrepo.UpdateUserAvatar(ctx, userID, avatarURL)
}

type defaultAuthProvider struct{}

func (defaultAuthProvider) HashPassword(password string) (string, error) {
	return auth.HashPassword(password)
}

func (defaultAuthProvider) CheckPassword(hashedPassword, password string) error {
	return auth.CheckPassword(hashedPassword, password)
}

func (defaultAuthProvider) GenerateTokenPair(userID uint) (accessToken, refreshToken string, err error) {
	return auth.GenerateTokenPair(userID)
}

func (defaultAuthProvider) RefreshTokens(refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	return auth.RefreshTokens(refreshToken)
}

type defaultUploadProvider struct{}

func (defaultUploadProvider) PrepareAvatar(userID uint, originalFilename string) (savePath, avatarURL string, err error) {
	return upload.PrepareAvatar(userID, originalFilename)
}

func (defaultUploadProvider) SaveFile(file *multipart.FileHeader, savePath string) error {
	return upload.SaveFile(file, savePath)
}

func (defaultUploadProvider) RemoveAvatar(avatarURL string) error {
	return upload.RemoveAvatar(avatarURL)
}

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
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", "", ErrPasswordWrong
		}
		return "", "", err
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
