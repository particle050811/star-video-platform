package interaction

import (
	"context"
	"time"
	dbdal "video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	rdbdal "video-platform/biz/dal/rdb"
)

type LikedVideoListResult struct {
	Items      []model.Video
	Total      int64
	NextCursor uint
	HasMore    bool
}

type UserCommentItem struct {
	ID        uint
	UserID    uint
	VideoID   uint
	Content   string
	LikeCount int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserCommentListResult struct {
	Items      []UserCommentItem
	Total      int64
	NextCursor uint
	HasMore    bool
}

type interactionDBStore interface {
	LikeVideo(ctx context.Context, userID, videoID uint) error
	CancelLikeVideo(ctx context.Context, userID, videoID uint) (bool, error)
	ListLikedVideoIDs(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.VideoLikeItem, int64, error)
	CreateComment(ctx context.Context, comment *model.Comment) error
	ListUserComments(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.UserComment, int64, bool, error)
	GetCommentByID(ctx context.Context, commentID uint) (*model.Comment, error)
	DeleteComment(ctx context.Context, commentID uint) error
}

type interactionVideoStore interface {
	GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error)
	ListVideosByIDs(ctx context.Context, videoIDs []uint) ([]model.Video, error)
}

type interactionCacheStore interface {
	DeleteVideoDetailCache(ctx context.Context, videoID uint) error
	BumpHotVideoCacheVersion(ctx context.Context) error
	BumpVideoCommentCacheVersion(ctx context.Context, videoID uint) error
}

type interactionStore struct {
	db     interactionDBStore
	videos interactionVideoStore
	cache  interactionCacheStore
}

type defaultInteractionCacheStore struct{}

func (defaultInteractionCacheStore) DeleteVideoDetailCache(ctx context.Context, videoID uint) error {
	return rdbdal.DeleteVideoDetailCache(ctx, videoID)
}

func (defaultInteractionCacheStore) BumpHotVideoCacheVersion(ctx context.Context) error {
	return rdbdal.BumpHotVideoCacheVersion(ctx)
}

func (defaultInteractionCacheStore) BumpVideoCommentCacheVersion(ctx context.Context, videoID uint) error {
	return rdbdal.BumpVideoCommentCacheVersion(ctx, videoID)
}

var interactions = interactionStore{
	db:     dbdal.Interactions,
	videos: dbdal.Videos,
	cache:  defaultInteractionCacheStore{},
}

func LikeVideo(ctx context.Context, userID, videoID uint) error {
	return interactions.LikeVideo(ctx, userID, videoID)
}

func CancelLikeVideo(ctx context.Context, userID, videoID uint) (bool, error) {
	return interactions.CancelLikeVideo(ctx, userID, videoID)
}

func ListLikedVideos(ctx context.Context, userID uint, cursor uint, limit int) (*LikedVideoListResult, error) {
	return interactions.ListLikedVideos(ctx, userID, cursor, limit)
}

func CreateComment(ctx context.Context, comment *model.Comment) error {
	return interactions.CreateComment(ctx, comment)
}

func ListUserComments(ctx context.Context, userID uint, cursor uint, limit int) (*UserCommentListResult, error) {
	return interactions.ListUserComments(ctx, userID, cursor, limit)
}

func GetCommentByID(ctx context.Context, commentID uint) (*model.Comment, error) {
	return interactions.GetCommentByID(ctx, commentID)
}

func DeleteComment(ctx context.Context, commentID uint) error {
	return interactions.DeleteComment(ctx, commentID)
}

func (s interactionStore) LikeVideo(ctx context.Context, userID, videoID uint) error {
	if err := s.db.LikeVideo(ctx, userID, videoID); err != nil {
		return err
	}
	s.deleteVideoCaches(ctx, videoID)
	return nil
}

func (s interactionStore) CancelLikeVideo(ctx context.Context, userID, videoID uint) (bool, error) {
	deleted, err := s.db.CancelLikeVideo(ctx, userID, videoID)
	if err != nil {
		return false, err
	}
	if deleted {
		s.deleteVideoCaches(ctx, videoID)
	}
	return deleted, nil
}

func (s interactionStore) ListLikedVideos(ctx context.Context, userID uint, cursor uint, limit int) (*LikedVideoListResult, error) {
	items, total, err := s.db.ListLikedVideoIDs(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	videoIDs := make([]uint, 0, len(items))
	for _, item := range items {
		videoIDs = append(videoIDs, item.VideoID)
	}

	videos, err := s.videos.ListVideosByIDs(ctx, videoIDs)
	if err != nil {
		return nil, err
	}

	nextCursor := uint(0)
	if hasMore && len(items) > 0 {
		nextCursor = items[len(items)-1].ID
	}

	return &LikedVideoListResult{
		Items:      videos,
		Total:      total,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s interactionStore) CreateComment(ctx context.Context, comment *model.Comment) error {
	if err := s.db.CreateComment(ctx, comment); err != nil {
		return err
	}
	s.deleteCommentCaches(ctx, comment.VideoID)
	return nil
}

func (s interactionStore) ListUserComments(ctx context.Context, userID uint, cursor uint, limit int) (*UserCommentListResult, error) {
	items, total, hasMore, err := s.db.ListUserComments(ctx, userID, cursor, limit)
	if err != nil {
		return nil, err
	}

	resultItems := make([]UserCommentItem, 0, len(items))
	for _, item := range items {
		resultItems = append(resultItems, UserCommentItem{
			ID:        item.ID,
			UserID:    item.UserID,
			VideoID:   item.VideoID,
			Content:   item.Content,
			LikeCount: item.LikeCount,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	nextCursor := uint(0)
	if len(resultItems) > 0 {
		nextCursor = resultItems[len(resultItems)-1].ID
	}

	return &UserCommentListResult{
		Items:      resultItems,
		Total:      total,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s interactionStore) GetCommentByID(ctx context.Context, commentID uint) (*model.Comment, error) {
	return s.db.GetCommentByID(ctx, commentID)
}

func (s interactionStore) DeleteComment(ctx context.Context, commentID uint) error {
	comment, err := s.db.GetCommentByID(ctx, commentID)
	if err != nil {
		return err
	}

	if err := s.db.DeleteComment(ctx, commentID); err != nil {
		return err
	}
	s.deleteCommentCaches(ctx, comment.VideoID)
	return nil
}

func (s interactionStore) deleteVideoCaches(ctx context.Context, videoID uint) {
	_ = s.cache.DeleteVideoDetailCache(ctx, videoID)
	_ = s.cache.BumpHotVideoCacheVersion(ctx)
}

func (s interactionStore) deleteCommentCaches(ctx context.Context, videoID uint) {
	s.deleteVideoCaches(ctx, videoID)
	_ = s.cache.BumpVideoCommentCacheVersion(ctx, videoID)
}
