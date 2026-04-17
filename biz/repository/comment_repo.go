package repository

import (
	"context"
	"time"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/rdb"
)

type VideoComment struct {
	ID        uint
	UserID    uint
	Content   string
	LikeCount int64
	CreatedAt time.Time
}

func ListVideoComments(ctx context.Context, videoID uint, offset, limit int) ([]VideoComment, int64, error) {
	version, err := rdb.GetVideoCommentCacheVersion(ctx, videoID)
	if err != nil {
		return nil, 0, err
	}

	var cached videoCommentCachePayload
	if ok, err := rdb.GetVideoCommentCache(ctx, videoID, version, offset, limit, &cached); err == nil && ok {
		return cached.Items, cached.Total, nil
	}

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

	_ = rdb.SetVideoCommentCache(ctx, videoID, version, offset, limit, videoCommentCachePayload{
		Items: items,
		Total: total,
	})
	return items, total, nil
}

type videoCommentCachePayload struct {
	Items []VideoComment `json:"items"`
	Total int64          `json:"total"`
}

func DeleteVideoCommentListCache(ctx context.Context, videoID uint) {
	_ = rdb.BumpVideoCommentCacheVersion(ctx, videoID)
}
