package db

import (
	"context"
	"video-platform/biz/dal/model"

	"gorm.io/gorm"
)

type SocialUser struct {
	ID        uint
	Username  string
	AvatarURL string
}

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

func ListFollowings(ctx context.Context, userID uint, offset, limit int) ([]SocialUser, int64, error) {
	var total int64
	if err := DB.WithContext(ctx).Model(&model.Relation{}).
		Where("from_user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	users := make([]SocialUser, 0)
	if total == 0 {
		return users, 0, nil
	}

	var toUserIDs []uint
	err := DB.WithContext(ctx).Model(&model.Relation{}).
		Where("from_user_id = ?", userID).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Pluck("to_user_id", &toUserIDs).Error
	if err != nil {
		return nil, 0, err
	}

	return listSocialUsersByIDs(ctx, toUserIDs, total)
}

func ListFollowers(ctx context.Context, userID uint, offset, limit int) ([]SocialUser, int64, error) {
	var total int64
	if err := DB.WithContext(ctx).Model(&model.Relation{}).
		Where("to_user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	users := make([]SocialUser, 0)
	if total == 0 {
		return users, 0, nil
	}

	var fromUserIDs []uint
	err := DB.WithContext(ctx).Model(&model.Relation{}).
		Where("to_user_id = ?", userID).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Pluck("from_user_id", &fromUserIDs).Error
	if err != nil {
		return nil, 0, err
	}

	return listSocialUsersByIDs(ctx, fromUserIDs, total)
}

func ListFriends(ctx context.Context, userID uint, offset, limit int) ([]SocialUser, int64, error) {
	baseQuery := DB.WithContext(ctx).
		Table("relations AS r1").
		Joins("JOIN relations AS r2 ON r1.to_user_id = r2.from_user_id AND r1.from_user_id = r2.to_user_id").
		Where("r1.from_user_id = ?", userID)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	users := make([]SocialUser, 0)
	if total == 0 {
		return users, 0, nil
	}

	var friendIDs []uint
	err := DB.WithContext(ctx).
		Table("relations AS r1").
		Select("r1.to_user_id").
		Joins("JOIN relations AS r2 ON r1.to_user_id = r2.from_user_id AND r1.from_user_id = r2.to_user_id").
		Where("r1.from_user_id = ?", userID).
		Order("r1.id DESC").
		Offset(offset).
		Limit(limit).
		Scan(&friendIDs).Error
	if err != nil {
		return nil, 0, err
	}

	return listSocialUsersByIDs(ctx, friendIDs, total)
}

func listSocialUsersByIDs(ctx context.Context, userIDs []uint, total int64) ([]SocialUser, int64, error) {
	userSnapshots, err := ListUserSnapshotsByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	users := make([]SocialUser, 0, len(userSnapshots))
	for _, user := range userSnapshots {
		users = append(users, SocialUser{
			ID:        user.ID,
			Username:  user.Username,
			AvatarURL: user.AvatarURL,
		})
	}

	return users, total, nil
}
