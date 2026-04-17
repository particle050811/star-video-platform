package repository

import (
	"context"
	"video-platform/biz/dal/model"
	dbdal "video-platform/biz/dal/db"
	rdbdal "video-platform/biz/dal/rdb"
	"video-platform/pkg/constant"
)

type userDBStore interface {
	ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error)
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByID(ctx context.Context, userID uint) (*model.User, error)
	UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error
	ListUsersByIDs(ctx context.Context, userIDs []uint) ([]model.User, error)
}

type userCacheStore interface {
	GetUserProfileCache(ctx context.Context, userID uint) (*rdbdal.UserProfileCache, bool, error)
	SetUserProfileCache(ctx context.Context, userID uint, value any) error
	DeleteUserProfileCache(ctx context.Context, userID uint) error
}

type userStore struct {
	db    userDBStore
	cache userCacheStore
}

var users = userStore{
	db:    dbdal.Users,
	cache: rdbdal.DefaultUserCache,
}

type UserProfile struct {
	ID             uint
	Username       string
	AvatarURL      string
	FollowingCount int64
	FollowerCount  int64
}

func ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error) {
	return users.ListUserIDsByUsername(ctx, username)
}

func CreateUser(ctx context.Context, user *model.User) error {
	return users.CreateUser(ctx, user)
}

func GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return users.GetUserByUsername(ctx, username)
}

func GetUserByID(ctx context.Context, userID uint) (*UserProfile, error) {
	return users.GetUserByID(ctx, userID)
}

func UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	return users.UpdateUserAvatar(ctx, userID, avatarURL)
}

func ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]UserProfile, error) {
	return users.ListUserSnapshotsByIDs(ctx, userIDs)
}

func (s userStore) ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error) {
	return s.db.ListUserIDsByUsername(ctx, username)
}

func (s userStore) CreateUser(ctx context.Context, user *model.User) error {
	return s.db.CreateUser(ctx, user)
}

func (s userStore) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.db.GetUserByUsername(ctx, username)
}

func (s userStore) GetUserByID(ctx context.Context, userID uint) (*UserProfile, error) {
	if cached, ok, err := s.cache.GetUserProfileCache(ctx, userID); err == nil && ok {
		return cachedUserToProfile(cached), nil
	}

	fetched, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetUserProfileCache(ctx, userID, newUserCachePayload(fetched))
	return newUserProfile(fetched), nil
}

func (s userStore) UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	if err := s.db.UpdateUserAvatar(ctx, userID, avatarURL); err != nil {
		return err
	}

	_ = s.cache.DeleteUserProfileCache(ctx, userID)
	return nil
}

func (s userStore) ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]UserProfile, error) {
	userMap := make(map[uint]UserProfile, len(userIDs))
	missedUserIDs := make([]uint, 0, len(userIDs))

	for _, userID := range userIDs {
		if user, ok, err := s.cache.GetUserProfileCache(ctx, userID); err == nil && ok {
			userMap[userID] = UserProfile{
				ID:             user.ID,
				Username:       user.Username,
				AvatarURL:      user.AvatarURL,
				FollowingCount: user.FollowingCount,
				FollowerCount:  user.FollowerCount,
			}
			continue
		}

		missedUserIDs = append(missedUserIDs, userID)
	}

	if len(missedUserIDs) > 0 {
		users, err := s.db.ListUsersByIDs(ctx, missedUserIDs)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			userMap[user.ID] = UserProfile{
				ID:             user.ID,
				Username:       user.Username,
				AvatarURL:      user.AvatarURL,
				FollowingCount: user.FollowingCount,
				FollowerCount:  user.FollowerCount,
			}
			_ = s.cache.SetUserProfileCache(ctx, user.ID, newUserCachePayload(&user))
		}
	}

	snapshots := make([]UserProfile, 0, len(userIDs))
	for _, userID := range userIDs {
		user, ok := userMap[userID]
		if !ok {
			user = UserProfile{
				ID:        userID,
				Username:  constant.DeletedUserName,
				AvatarURL: "",
			}
		}
		snapshots = append(snapshots, user)
	}

	return snapshots, nil
}

func newUserCachePayload(user *model.User) rdbdal.UserProfileCache {
	return rdbdal.UserProfileCache{
		ID:             user.ID,
		Username:       user.Username,
		AvatarURL:      user.AvatarURL,
		FollowingCount: user.FollowingCount,
		FollowerCount:  user.FollowerCount,
	}
}

func newUserProfile(user *model.User) *UserProfile {
	return &UserProfile{
		ID:             user.ID,
		Username:       user.Username,
		AvatarURL:      user.AvatarURL,
		FollowingCount: user.FollowingCount,
		FollowerCount:  user.FollowerCount,
	}
}

func cachedUserToProfile(u *rdbdal.UserProfileCache) *UserProfile {
	return &UserProfile{
		ID:             u.ID,
		Username:       u.Username,
		AvatarURL:      u.AvatarURL,
		FollowingCount: u.FollowingCount,
		FollowerCount:  u.FollowerCount,
	}
}
