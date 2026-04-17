package db

import (
	"context"
	"time"
	"video-platform/biz/dal/model"
)

type VideoComment struct {
	ID        uint
	UserID    uint
	Content   string
	LikeCount int64
	CreatedAt time.Time
}

func ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) ([]VideoComment, int64, bool, error) {
	var total int64
	if err := DB.WithContext(ctx).Model(&model.Comment{}).
		Where("video_id = ?", videoID).
		Count(&total).Error; err != nil {
		return nil, 0, false, err
	}

	comments := make([]VideoComment, 0)
	if total == 0 {
		return comments, 0, false, nil
	}

	query := DB.WithContext(ctx).
		Model(&model.Comment{}).
		Select("id, user_id, content, like_count, created_at").
		Where("video_id = ?", videoID)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}

	err := query.
		Order("id DESC").
		Limit(limit + 1).
		Scan(&comments).Error
	if err != nil {
		return nil, 0, false, err
	}

	hasMore := len(comments) > limit
	if hasMore {
		comments = comments[:limit]
	}

	return comments, total, hasMore, nil
}
