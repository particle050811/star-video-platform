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

type commentDBStore interface {
	ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) ([]db.VideoComment, int64, bool, error)
}

type commentCacheStore interface {
	GetVideoCommentCacheVersion(ctx context.Context, videoID uint) (int64, error)
	GetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, dest any) (bool, error)
	SetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, value any) error
	BumpVideoCommentCacheVersion(ctx context.Context, videoID uint) error
}

type commentStore struct {
	db    commentDBStore
	cache commentCacheStore
}

type defaultCommentDBStore struct{}

func (defaultCommentDBStore) ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) ([]db.VideoComment, int64, bool, error) {
	return db.ListVideoComments(ctx, videoID, cursor, limit)
}

type defaultCommentCacheStore struct{}

func (defaultCommentCacheStore) GetVideoCommentCacheVersion(ctx context.Context, videoID uint) (int64, error) {
	return rdb.GetVideoCommentCacheVersion(ctx, videoID)
}

func (defaultCommentCacheStore) GetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
	return rdb.GetVideoCommentCache(ctx, videoID, version, cursor, limit, dest)
}

func (defaultCommentCacheStore) SetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, value any) error {
	return rdb.SetVideoCommentCache(ctx, videoID, version, cursor, limit, value)
}

func (defaultCommentCacheStore) BumpVideoCommentCacheVersion(ctx context.Context, videoID uint) error {
	return rdb.BumpVideoCommentCacheVersion(ctx, videoID)
}

var comments = commentStore{
	db:    defaultCommentDBStore{},
	cache: defaultCommentCacheStore{},
}

func ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) (*VideoCommentListResult, error) {
	return comments.ListVideoComments(ctx, videoID, cursor, limit)
}

func (s commentStore) ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) (*VideoCommentListResult, error) {
	version, err := s.cache.GetVideoCommentCacheVersion(ctx, videoID)
	if err != nil {
		return nil, err
	}

	var cached videoCommentCachePayload
	if ok, err := s.cache.GetVideoCommentCache(ctx, videoID, version, cursor, limit, &cached); err == nil && ok {
		return &VideoCommentListResult{
			Items:      cached.Items,
			Total:      cached.Total,
			NextCursor: cached.NextCursor,
			HasMore:    cached.HasMore,
		}, nil
	}

	comments, total, hasMore, err := s.db.ListVideoComments(ctx, videoID, cursor, limit)
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
	_ = s.cache.SetVideoCommentCache(ctx, videoID, version, cursor, limit, videoCommentCachePayload{
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
	_ = comments.cache.BumpVideoCommentCacheVersion(ctx, videoID)
}
