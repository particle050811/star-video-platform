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

type VideoCommentListResult struct {
	Items      []VideoComment
	Total      int64
	NextCursor uint
	HasMore    bool
}

func ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) (*VideoCommentListResult, error) {
	version, err := rdb.GetVideoCommentCacheVersion(ctx, videoID)
	if err != nil {
		return nil, err
	}

	var cached videoCommentCachePayload
	if ok, err := rdb.GetVideoCommentCache(ctx, videoID, version, cursor, limit, &cached); err == nil && ok {
		return &VideoCommentListResult{
			Items:      cached.Items,
			Total:      cached.Total,
			NextCursor: cached.NextCursor,
			HasMore:    cached.HasMore,
		}, nil
	}

	comments, total, hasMore, err := db.ListVideoComments(ctx, videoID, cursor, limit)
	if err != nil {
		return nil, err
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

	nextCursor := uint(0)
	if len(items) > 0 {
		nextCursor = items[len(items)-1].ID
	}

	result := &VideoCommentListResult{
		Items:      items,
		Total:      total,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
	_ = rdb.SetVideoCommentCache(ctx, videoID, version, cursor, limit, videoCommentCachePayload{
		Items:      items,
		Total:      total,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
	return result, nil
}

type videoCommentCachePayload struct {
	Items      []VideoComment `json:"items"`
	Total      int64          `json:"total"`
	NextCursor uint           `json:"next_cursor"`
	HasMore    bool           `json:"has_more"`
}

func DeleteVideoCommentListCache(ctx context.Context, videoID uint) {
	_ = rdb.BumpVideoCommentCacheVersion(ctx, videoID)
}
