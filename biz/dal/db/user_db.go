package db

import (
	"context"
	"video-platform/biz/dal/model"

	"gorm.io/gorm"
)

type UserDB struct {
	db *gorm.DB
}

func NewUserDB(gdb *gorm.DB) UserDB {
	return UserDB{db: gdb}
}

var Users = NewUserDB(DB)

func (u UserDB) gormDB() *gorm.DB {
	if u.db != nil {
		return u.db
	}
	return DB
}

func (u UserDB) CreateUser(ctx context.Context, user *model.User) error {
	return u.gormDB().WithContext(ctx).Create(user).Error
}

func (u UserDB) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := u.gormDB().WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u UserDB) ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error) {
	var userIDs []uint
	if err := u.gormDB().WithContext(ctx).Model(&model.User{}).
		Where("username LIKE ?", "%"+username+"%").
		Pluck("id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (u UserDB) ListUsersByIDs(ctx context.Context, userIDs []uint) ([]model.User, error) {
	users := make([]model.User, 0, len(userIDs))
	if len(userIDs) == 0 {
		return users, nil
	}

	if err := u.gormDB().WithContext(ctx).
		Where("id IN ?", userIDs).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (u UserDB) GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	if err := u.gormDB().WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u UserDB) UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	tx := u.gormDB().WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("avatar_url", avatarURL)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func CreateUser(ctx context.Context, user *model.User) error {
	return Users.CreateUser(ctx, user)
}

func GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return Users.GetUserByUsername(ctx, username)
}

func ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error) {
	return Users.ListUserIDsByUsername(ctx, username)
}

func ListUsersByIDs(ctx context.Context, userIDs []uint) ([]model.User, error) {
	return Users.ListUsersByIDs(ctx, userIDs)
}

func GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	return Users.GetUserByID(ctx, userID)
}

func UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	return Users.UpdateUserAvatar(ctx, userID, avatarURL)
}
