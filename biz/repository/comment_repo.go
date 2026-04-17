package repository

import (
	"context"
	"time"
	"video-platform/biz/dal/db"
)

type VideoComment struct {
	ID        uint
	UserID    uint
	Content   string
	LikeCount int64
	CreatedAt time.Time
}

func ListVideoComments(ctx context.Context, videoID uint, offset, limit int) ([]VideoComment, int64, error) {
	comments, total, err := db.ListVideoComments(ctx, videoID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	items := make([]VideoComment, 0, len(comments))
	for _, comment := range comments {
		items = append(items, VideoComment{
			ID:        comment.ID,
			UserID:    comment.UserID,
			Content:   comment.Content,
			LikeCount: comment.LikeCount,
			CreatedAt: comment.CreatedAt,
		})
	}

	return items, total, nil
}
