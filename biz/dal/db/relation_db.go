package db

import (
	"context"
	"video-platform/biz/dal/model"

	"gorm.io/gorm"
)

func FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	var deleted bool

	err := DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func ListFollowingIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error) {
	var total int64
	if err := DB.WithContext(ctx).Model(&model.Relation{}).
		Where("from_user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	userIDs := make([]uint, 0)
	if total == 0 {
		return userIDs, 0, nil
	}

	err := DB.WithContext(ctx).Model(&model.Relation{}).
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

func ListFollowerIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error) {
	var total int64
	if err := DB.WithContext(ctx).Model(&model.Relation{}).
		Where("to_user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	userIDs := make([]uint, 0)
	if total == 0 {
		return userIDs, 0, nil
	}

	err := DB.WithContext(ctx).Model(&model.Relation{}).
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

func ListFriendIDs(ctx context.Context, userID uint, offset, limit int) ([]uint, int64, error) {
	baseQuery := DB.WithContext(ctx).
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

	err := DB.WithContext(ctx).
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
