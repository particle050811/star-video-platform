package db

import (
	"context"
	"errors"
	"time"
	"video-platform/biz/dal/model"

	"gorm.io/gorm"
)

type UserComment struct {
	ID        uint
	UserID    uint
	VideoID   uint
	Content   string
	LikeCount int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type VideoLikeItem struct {
	ID      uint
	VideoID uint
}

type InteractionDB struct {
	db *gorm.DB
}

func NewInteractionDB(gdb *gorm.DB) InteractionDB {
	return InteractionDB{db: gdb}
}

var Interactions = NewInteractionDB(DB)

func (i InteractionDB) gormDB() *gorm.DB {
	if i.db != nil {
		return i.db
	}
	return DB
}

func (i InteractionDB) LikeVideo(ctx context.Context, userID, videoID uint) error {
	return i.gormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.VideoLike{
			UserID:  userID,
			VideoID: videoID,
		}).Error; err != nil {
			return err
		}

		res := tx.Model(&model.Video{}).
			Where("id = ?", videoID).
			Update("like_count", gorm.Expr("like_count + ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (i InteractionDB) CancelLikeVideo(ctx context.Context, userID, videoID uint) (bool, error) {
	var deleted bool

	err := i.gormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("user_id = ? AND video_id = ?", userID, videoID).
			Delete(&model.VideoLike{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		deleted = true

		res = tx.Model(&model.Video{}).
			Where("id = ?", videoID).
			Update("like_count", gorm.Expr("like_count - ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return deleted, nil
}

func (i InteractionDB) ListLikedVideoIDs(ctx context.Context, userID uint, cursor uint, limit int) ([]VideoLikeItem, int64, error) {
	baseQuery := i.gormDB().WithContext(ctx).
		Table("video_likes").
		Joins("JOIN videos ON videos.id = video_likes.video_id AND videos.deleted_at IS NULL").
		Where("video_likes.user_id = ?", userID)

	var total int64
	if err := baseQuery.
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]VideoLikeItem, 0)
	if total == 0 {
		return items, 0, nil
	}

	query := i.gormDB().WithContext(ctx).
		Table("video_likes").
		Select("video_likes.id, video_likes.video_id").
		Joins("JOIN videos ON videos.id = video_likes.video_id AND videos.deleted_at IS NULL").
		Where("video_likes.user_id = ?", userID)
	if cursor > 0 {
		query = query.Where("video_likes.id < ?", cursor)
	}

	if err := query.Order("video_likes.id DESC").Limit(limit).Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (i InteractionDB) CreateComment(ctx context.Context, comment *model.Comment) error {
	return i.gormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		res := tx.Model(&model.Video{}).
			Where("id = ?", comment.VideoID).
			Update("comment_count", gorm.Expr("comment_count + ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (i InteractionDB) ListUserComments(ctx context.Context, userID uint, cursor uint, limit int) ([]UserComment, int64, bool, error) {
	var total int64
	if err := i.gormDB().WithContext(ctx).Model(&model.Comment{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, false, err
	}

	items := make([]UserComment, 0)
	if total == 0 {
		return items, 0, false, nil
	}

	query := i.gormDB().WithContext(ctx).
		Model(&model.Comment{}).
		Select("id, user_id, video_id, content, like_count, created_at, updated_at").
		Where("user_id = ?", userID)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}

	err := query.Order("id DESC").Limit(limit + 1).Scan(&items).Error
	if err != nil {
		return nil, 0, false, err
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	return items, total, hasMore, nil
}

func (i InteractionDB) GetCommentByID(ctx context.Context, commentID uint) (*model.Comment, error) {
	var comment model.Comment
	if err := i.gormDB().WithContext(ctx).First(&comment, commentID).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (i InteractionDB) DeleteComment(ctx context.Context, commentID uint) error {
	return i.gormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var comment model.Comment
		if err := tx.First(&comment, commentID).Error; err != nil {
			return err
		}

		res := tx.Delete(&model.Comment{}, commentID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		res = tx.Model(&model.Video{}).
			Where("id = ?", comment.VideoID).
			Update("comment_count", gorm.Expr("comment_count - ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func LikeVideo(ctx context.Context, userID, videoID uint) error {
	return Interactions.LikeVideo(ctx, userID, videoID)
}

func CancelLikeVideo(ctx context.Context, userID, videoID uint) (bool, error) {
	return Interactions.CancelLikeVideo(ctx, userID, videoID)
}

func ListLikedVideoIDs(ctx context.Context, userID uint, cursor uint, limit int) ([]VideoLikeItem, int64, error) {
	return Interactions.ListLikedVideoIDs(ctx, userID, cursor, limit)
}

func CreateComment(ctx context.Context, comment *model.Comment) error {
	return Interactions.CreateComment(ctx, comment)
}

func ListUserComments(ctx context.Context, userID uint, cursor uint, limit int) ([]UserComment, int64, bool, error) {
	return Interactions.ListUserComments(ctx, userID, cursor, limit)
}

func GetCommentByID(ctx context.Context, commentID uint) (*model.Comment, error) {
	return Interactions.GetCommentByID(ctx, commentID)
}

func DeleteComment(ctx context.Context, commentID uint) error {
	return Interactions.DeleteComment(ctx, commentID)
}
