package db

import (
	"context"
	"video-platform/biz/dal/model"

	"gorm.io/gorm"
)

type RelationDB struct {
	db *gorm.DB
}

func NewRelationDB(gdb *gorm.DB) RelationDB {
	return RelationDB{db: gdb}
}

var Relations = NewRelationDB(DB)

func (r RelationDB) gormDB() *gorm.DB {
	if r.db != nil {
		return r.db
	}
	return DB
}

func (r RelationDB) FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	return r.gormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.Relation{
			FromUserID: fromUserID,
			ToUserID:   toUserID,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.User{}).
			Where("id = ?", fromUserID).
			Update("following_count", gorm.Expr("following_count + ?", 1)).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.User{}).
			Where("id = ?", toUserID).
			Update("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r RelationDB) UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	var deleted bool

	err := r.gormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("from_user_id = ? AND to_user_id = ?", fromUserID, toUserID).
			Delete(&model.Relation{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		deleted = true

		if err := tx.Model(&model.User{}).
			Where("id = ?", fromUserID).
			Update("following_count", gorm.Expr("following_count - ?", 1)).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.User{}).
			Where("id = ?", toUserID).
			Update("follower_count", gorm.Expr("follower_count - ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})

	return deleted, err
}

func (r RelationDB) ListFollowingIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error) {
	var total int64
	if err := r.gormDB().WithContext(ctx).Model(&model.Relation{}).
		Where("from_user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	userIDs := make([]uint, 0)
	if total == 0 {
		return userIDs, 0, nil
	}

	err := r.gormDB().WithContext(ctx).Model(&model.Relation{}).
		Where("from_user_id = ?", userID).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Pluck("to_user_id", &userIDs).Error
	if err != nil {
		return nil, 0, err
	}

	return userIDs, total, nil
}

func (r RelationDB) ListFollowerIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error) {
	var total int64
	if err := r.gormDB().WithContext(ctx).Model(&model.Relation{}).
		Where("to_user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	userIDs := make([]uint, 0)
	if total == 0 {
		return userIDs, 0, nil
	}

	err := r.gormDB().WithContext(ctx).Model(&model.Relation{}).
		Where("to_user_id = ?", userID).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Pluck("from_user_id", &userIDs).Error
	if err != nil {
		return nil, 0, err
	}

	return userIDs, total, nil
}

func (r RelationDB) ListFriendIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error) {
	baseQuery := r.gormDB().WithContext(ctx).
		Table("relations AS r1").
		Joins("JOIN relations AS r2 ON r1.to_user_id = r2.from_user_id AND r1.from_user_id = r2.to_user_id").
		Where("r1.from_user_id = ?", userID)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	userIDs := make([]uint, 0)
	if total == 0 {
		return userIDs, 0, nil
	}

	err := r.gormDB().WithContext(ctx).
		Table("relations AS r1").
		Joins("JOIN relations AS r2 ON r1.to_user_id = r2.from_user_id AND r1.from_user_id = r2.to_user_id").
		Where("r1.from_user_id = ?", userID).
		Order("r1.id DESC").
		Offset(offset).
		Limit(limit).
		Pluck("r1.to_user_id", &userIDs).Error
	if err != nil {
		return nil, 0, err
	}

	return userIDs, total, nil
}

func FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	return Relations.FollowUser(ctx, fromUserID, toUserID)
}

func UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	return Relations.UnfollowUser(ctx, fromUserID, toUserID)
}

func ListFollowingIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error) {
	return Relations.ListFollowingIDs(ctx, userID, offset, limit)
}

func ListFollowerIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error) {
	return Relations.ListFollowerIDs(ctx, userID, offset, limit)
}

func ListFriendIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error) {
	return Relations.ListFriendIDs(ctx, userID, offset, limit)
}
