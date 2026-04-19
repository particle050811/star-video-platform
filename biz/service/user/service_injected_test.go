package user

import (
	"context"
	"errors"
	"testing"
	"video-platform/biz/dal/model"
	userrepo "video-platform/biz/repository/user"
	"video-platform/pkg/auth"

	"gorm.io/gorm"
)

type fakeUserRepository struct {
	createUserFn        func(ctx context.Context, user *model.User) error
	getUserByUsernameFn func(ctx context.Context, username string) (*model.User, error)
	getUserByIDFn       func(ctx context.Context, userID uint) (*userrepo.UserProfile, error)
	updateUserAvatarFn  func(ctx context.Context, userID uint, avatarURL string) error
}

func (f fakeUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	return f.createUserFn(ctx, user)
}

func (f fakeUserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return f.getUserByUsernameFn(ctx, username)
}

func (f fakeUserRepository) GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
	return f.getUserByIDFn(ctx, userID)
}

func (f fakeUserRepository) UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	return f.updateUserAvatarFn(ctx, userID, avatarURL)
}

type fakeAuthProvider struct {
	hashPasswordFn      func(password string) (string, error)
	checkPasswordFn     func(hashedPassword, password string) error
	generateTokenPairFn func(userID uint) (accessToken, refreshToken string, err error)
	refreshTokensFn     func(refreshToken string) (newAccessToken, newRefreshToken string, err error)
}

func (f fakeAuthProvider) HashPassword(password string) (string, error) {
	return f.hashPasswordFn(password)
}

func (f fakeAuthProvider) CheckPassword(hashedPassword, password string) error {
	return f.checkPasswordFn(hashedPassword, password)
}

func (f fakeAuthProvider) GenerateTokenPair(userID uint) (accessToken, refreshToken string, err error) {
	return f.generateTokenPairFn(userID)
}

func (f fakeAuthProvider) RefreshTokens(refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	return f.refreshTokensFn(refreshToken)
}

func TestUserServiceRegisterMapsDuplicatedKey(t *testing.T) {
	svc := userService{
		repo: fakeUserRepository{
			createUserFn: func(ctx context.Context, user *model.User) error {
				if user.Username != "alice" {
					t.Fatalf("expected username %q, got %q", "alice", user.Username)
				}
				if user.Password != "hashed-password" {
					t.Fatalf("expected password %q, got %q", "hashed-password", user.Password)
				}
				return gorm.ErrDuplicatedKey
			},
		},
		auth: fakeAuthProvider{
			hashPasswordFn: func(password string) (string, error) {
				if password != "plain-password" {
					t.Fatalf("expected password %q, got %q", "plain-password", password)
				}
				return "hashed-password", nil
			},
		},
	}

	err := svc.Register(context.Background(), "alice", "plain-password")
	if !errors.Is(err, ErrUserExists) {
		t.Fatalf("expected error %v, got %v", ErrUserExists, err)
	}
}

func TestUserServiceLoginSuccess(t *testing.T) {
	svc := userService{
		repo: fakeUserRepository{
			getUserByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
				if username != "alice" {
					t.Fatalf("expected username %q, got %q", "alice", username)
				}
				return &model.User{
					ID:       42,
					Username: "alice",
					Password: "hashed-password",
				}, nil
			},
		},
		auth: fakeAuthProvider{
			checkPasswordFn: func(hashedPassword, password string) error {
				if hashedPassword != "hashed-password" {
					t.Fatalf("expected hashed password %q, got %q", "hashed-password", hashedPassword)
				}
				if password != "plain-password" {
					t.Fatalf("expected password %q, got %q", "plain-password", password)
				}
				return nil
			},
			generateTokenPairFn: func(userID uint) (string, string, error) {
				if userID != 42 {
					t.Fatalf("expected user id %d, got %d", 42, userID)
				}
				return "access-token", "refresh-token", nil
			},
		},
	}

	accessToken, refreshToken, err := svc.Login(context.Background(), "alice", "plain-password")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if accessToken != "access-token" {
		t.Fatalf("expected access token %q, got %q", "access-token", accessToken)
	}
	if refreshToken != "refresh-token" {
		t.Fatalf("expected refresh token %q, got %q", "refresh-token", refreshToken)
	}
}

func TestUserServiceRefreshTokenMapsKnownErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   error
		wantErr error
	}{
		{
			name:    "expired token",
			input:   auth.ErrTokenExpired,
			wantErr: ErrTokenExpired,
		},
		{
			name:    "invalid token",
			input:   auth.ErrTokenInvalid,
			wantErr: ErrTokenInvalid,
		},
		{
			name:    "malformed token",
			input:   auth.ErrTokenMalformed,
			wantErr: ErrTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := userService{
				auth: fakeAuthProvider{
					refreshTokensFn: func(refreshToken string) (string, string, error) {
						if refreshToken != "refresh-token" {
							t.Fatalf("expected refresh token %q, got %q", "refresh-token", refreshToken)
						}
						return "", "", tt.input
					},
				},
			}

			_, _, err := svc.RefreshToken(context.Background(), "refresh-token")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestUserServiceGetUserInfoMapsRecordNotFound(t *testing.T) {
	svc := userService{
		repo: fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
				if userID != 99 {
					t.Fatalf("expected user id %d, got %d", 99, userID)
				}
				return nil, gorm.ErrRecordNotFound
			},
		},
	}

	_, err := svc.GetUserInfo(context.Background(), 99)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected error %v, got %v", ErrUserNotFound, err)
	}
}
