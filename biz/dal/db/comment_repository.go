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

func ListVideoComments(ctx context.Context, videoID uint, offset, limit int) ([]VideoComment, int64, error) {
	var total int64
	if err := DB.WithContext(ctx).Model(&model.Comment{}).
		Where("video_id = ?", videoID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	comments := make([]VideoComment, 0)
	if total == 0 {
		return comments, 0, nil
	}

	err := DB.WithContext(ctx).
		Model(&model.Comment{}).
		Select("id, user_id, content, like_count, created_at").
		Where("video_id = ?", videoID).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Scan(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}
