package user

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"video-platform/biz/dal/model"
	userrepo "video-platform/biz/repository/user"
	"video-platform/pkg/auth"
	"video-platform/pkg/upload"

	"golang.org/x/crypto/bcrypt"
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

type fakeUploadProvider struct {
	prepareAvatarFn func(userID uint, originalFilename string) (savePath, avatarURL string, err error)
	saveFileFn      func(file *multipart.FileHeader, savePath string) error
	removeAvatarFn  func(avatarURL string) error
}

func (f fakeUploadProvider) PrepareAvatar(userID uint, originalFilename string) (savePath, avatarURL string, err error) {
	return f.prepareAvatarFn(userID, originalFilename)
}

func (f fakeUploadProvider) SaveFile(file *multipart.FileHeader, savePath string) error {
	return f.saveFileFn(file, savePath)
}

func (f fakeUploadProvider) RemoveAvatar(avatarURL string) error {
	return f.removeAvatarFn(avatarURL)
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

func TestUserServiceLoginMapsPasswordMismatchOnly(t *testing.T) {
	svc := userService{
		repo: fakeUserRepository{
			getUserByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
				return &model.User{
					ID:       42,
					Username: username,
					Password: "hashed-password",
				}, nil
			},
		},
		auth: fakeAuthProvider{
			checkPasswordFn: func(hashedPassword, password string) error {
				return bcrypt.ErrMismatchedHashAndPassword
			},
		},
	}

	_, _, err := svc.Login(context.Background(), "alice", "wrong-password")
	if !errors.Is(err, ErrPasswordWrong) {
		t.Fatalf("expected error %v, got %v", ErrPasswordWrong, err)
	}
}

func TestUserServiceLoginReturnsInternalPasswordCheckError(t *testing.T) {
	wantErr := bcrypt.ErrHashTooShort
	svc := userService{
		repo: fakeUserRepository{
			getUserByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
				return &model.User{
					ID:       42,
					Username: username,
					Password: "bad-hash",
				}, nil
			},
		},
		auth: fakeAuthProvider{
			checkPasswordFn: func(hashedPassword, password string) error {
				return wantErr
			},
		},
	}

	_, _, err := svc.Login(context.Background(), "alice", "plain-password")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
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

func TestUserServiceUpdateUserAvatarRemovesPreparedAvatarOnSaveFailure(t *testing.T) {
	file := &multipart.FileHeader{Filename: "avatar.png"}
	saveErr := errors.New("save failed")
	var removed []string

	svc := userService{
		repo: fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
				return &userrepo.UserProfile{
					ID:        userID,
					Username:  "alice",
					AvatarURL: "/static/avatars/old.png",
				}, nil
			},
			updateUserAvatarFn: func(ctx context.Context, userID uint, avatarURL string) error {
				t.Fatal("repo update should not be called when save file fails")
				return nil
			},
		},
		upload: fakeUploadProvider{
			prepareAvatarFn: func(userID uint, originalFilename string) (string, string, error) {
				return "/tmp/new-avatar.png", "/static/avatars/new.png", nil
			},
			saveFileFn: func(gotFile *multipart.FileHeader, savePath string) error {
				if gotFile != file {
					t.Fatalf("expected same file pointer")
				}
				if savePath != "/tmp/new-avatar.png" {
					t.Fatalf("expected save path %q, got %q", "/tmp/new-avatar.png", savePath)
				}
				return saveErr
			},
			removeAvatarFn: func(avatarURL string) error {
				removed = append(removed, avatarURL)
				return nil
			},
		},
	}

	err := svc.UpdateUserAvatar(context.Background(), 7, file)
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected error %v, got %v", saveErr, err)
	}
	if len(removed) != 1 || removed[0] != "/static/avatars/new.png" {
		t.Fatalf("expected prepared avatar cleanup, got %v", removed)
	}
}

func TestUserServiceUpdateUserAvatarRemovesPreparedAvatarOnRepoFailure(t *testing.T) {
	file := &multipart.FileHeader{Filename: "avatar.png"}
	updateErr := errors.New("update failed")
	var removed []string

	svc := userService{
		repo: fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
				return &userrepo.UserProfile{
					ID:        userID,
					Username:  "alice",
					AvatarURL: "/static/avatars/old.png",
				}, nil
			},
			updateUserAvatarFn: func(ctx context.Context, userID uint, avatarURL string) error {
				if avatarURL != "/static/avatars/new.png" {
					t.Fatalf("expected avatar url %q, got %q", "/static/avatars/new.png", avatarURL)
				}
				return updateErr
			},
		},
		upload: fakeUploadProvider{
			prepareAvatarFn: func(userID uint, originalFilename string) (string, string, error) {
				return "/tmp/new-avatar.png", "/static/avatars/new.png", nil
			},
			saveFileFn: func(gotFile *multipart.FileHeader, savePath string) error {
				return nil
			},
			removeAvatarFn: func(avatarURL string) error {
				removed = append(removed, avatarURL)
				return nil
			},
		},
	}

	err := svc.UpdateUserAvatar(context.Background(), 7, file)
	if !errors.Is(err, updateErr) {
		t.Fatalf("expected error %v, got %v", updateErr, err)
	}
	if len(removed) != 1 || removed[0] != "/static/avatars/new.png" {
		t.Fatalf("expected prepared avatar cleanup, got %v", removed)
	}
}

func TestUserServiceUpdateUserAvatarRemovesOldAvatarAfterSuccess(t *testing.T) {
	file := &multipart.FileHeader{Filename: "avatar.png"}
	var removed []string

	svc := userService{
		repo: fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
				return &userrepo.UserProfile{
					ID:        userID,
					Username:  "alice",
					AvatarURL: "/static/avatars/old.png",
				}, nil
			},
			updateUserAvatarFn: func(ctx context.Context, userID uint, avatarURL string) error {
				return nil
			},
		},
		upload: fakeUploadProvider{
			prepareAvatarFn: func(userID uint, originalFilename string) (string, string, error) {
				return "/tmp/new-avatar.png", "/static/avatars/new.png", nil
			},
			saveFileFn: func(gotFile *multipart.FileHeader, savePath string) error {
				return nil
			},
			removeAvatarFn: func(avatarURL string) error {
				removed = append(removed, avatarURL)
				return nil
			},
		},
	}

	err := svc.UpdateUserAvatar(context.Background(), 7, file)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(removed) != 1 || removed[0] != "/static/avatars/old.png" {
		t.Fatalf("expected old avatar cleanup, got %v", removed)
	}
}

func TestUserServiceUpdateUserAvatarMapsUnsupportedExt(t *testing.T) {
	file := &multipart.FileHeader{Filename: "avatar.exe"}
	var removed []string

	svc := userService{
		repo: fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
				return &userrepo.UserProfile{
					ID:       userID,
					Username: "alice",
				}, nil
			},
		},
		upload: fakeUploadProvider{
			prepareAvatarFn: func(userID uint, originalFilename string) (string, string, error) {
				return "", "", upload.ErrUnsupportedAvatarExt
			},
			saveFileFn: func(gotFile *multipart.FileHeader, savePath string) error {
				t.Fatal("save file should not be called when prepare avatar fails")
				return nil
			},
			removeAvatarFn: func(avatarURL string) error {
				removed = append(removed, avatarURL)
				return nil
			},
		},
	}

	err := svc.UpdateUserAvatar(context.Background(), 7, file)
	if !errors.Is(err, ErrUnsupportedAvatarExt) {
		t.Fatalf("expected error %v, got %v", ErrUnsupportedAvatarExt, err)
	}
	if len(removed) != 1 || removed[0] != "" {
		t.Fatalf("expected empty prepared avatar cleanup, got %v", removed)
	}
}
