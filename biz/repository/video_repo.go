package repository

import (
	"context"
	dbdal "video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	rdbdal "video-platform/biz/dal/rdb"
	"video-platform/pkg/parser"
)

type videoDBStore interface {
	CreateVideo(ctx context.Context, video *model.Video) error
	GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error)
	ListVideosByUserID(ctx context.Context, userID uint, cursor uint, limit int) ([]model.Video, error)
	SearchVideos(ctx context.Context, params dbdal.VideoQuery) ([]model.Video, error)
	ListHotVideos(ctx context.Context, cursor parser.HotVideoCursorValue, limit int) ([]model.Video, error)
}

type videoCacheStore interface {
	BumpHotVideoCacheVersion(ctx context.Context) error
	GetVideoDetailCache(ctx context.Context, videoID uint, dest any) (bool, error)
	SetVideoDetailCache(ctx context.Context, videoID uint, value any) error
	GetHotVideoCacheVersion(ctx context.Context) (int64, error)
	GetHotVideoCache(ctx context.Context, version int64, cursor parser.HotVideoCursorValue, limit int, dest any) (bool, error)
	SetHotVideoCache(ctx context.Context, version int64, cursor parser.HotVideoCursorValue, limit int, value any) error
}

type videoStore struct {
	db    videoDBStore
	cache videoCacheStore
}

type VideoListResult struct {
	Items           []model.Video
	NextCursor      uint
	NextCursorToken string
	HasMore         bool
}

var videos = videoStore{
	db:    dbdal.Videos,
	cache: rdbdal.DefaultVideoCache,
}

func CreateVideo(ctx context.Context, video *model.Video) error {
	return videos.CreateVideo(ctx, video)
}

func GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return videos.GetVideoByID(ctx, videoID)
}

func ListVideosByUserID(ctx context.Context, userID uint, cursor uint, limit int) (*VideoListResult, error) {
	return videos.ListVideosByUserID(ctx, userID, cursor, limit)
}

func SearchVideos(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, cursor uint, limit int) (*VideoListResult, error) {
	return videos.SearchVideos(ctx, keywords, userIDs, fromDate, toDate, sortBy, cursor, limit)
}

func ListHotVideos(ctx context.Context, cursor parser.HotVideoCursorValue, limit int) (*VideoListResult, error) {
	return videos.ListHotVideos(ctx, cursor, limit)
}

func (s videoStore) CreateVideo(ctx context.Context, video *model.Video) error {
	if err := s.db.CreateVideo(ctx, video); err != nil {
		return err
	}

	_ = s.cache.BumpHotVideoCacheVersion(ctx)
	return nil
}

func (s videoStore) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	var video model.Video
	if ok, err := s.cache.GetVideoDetailCache(ctx, videoID, &video); err == nil && ok {
		return &video, nil
	}

	fetched, err := s.db.GetVideoByID(ctx, videoID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetVideoDetailCache(ctx, videoID, fetched)
	return fetched, nil
}

func (s videoStore) ListVideosByUserID(ctx context.Context, userID uint, cursor uint, limit int) (*VideoListResult, error) {
	videos, err := s.db.ListVideosByUserID(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	return buildVideoListResult(videos, cursor, limit), nil
}

func (s videoStore) SearchVideos(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, cursor uint, limit int) (*VideoListResult, error) {
	videos, err := s.db.SearchVideos(ctx, dbdal.VideoQuery{
		Keywords: keywords,
		UserIDs:  userIDs,
		FromDate: fromDate,
		ToDate:   toDate,
		SortBy:   sortBy,
		Cursor:   cursor,
		Limit:    limit + 1,
	})
	if err != nil {
		return nil, err
	}

	return buildVideoListResult(videos, cursor, limit), nil
}

func (s videoStore) ListHotVideos(ctx context.Context, cursor parser.HotVideoCursorValue, limit int) (*VideoListResult, error) {
	version, err := s.cache.GetHotVideoCacheVersion(ctx)
	if err != nil {
		return nil, err
	}

	var result VideoListResult
	if ok, err := s.cache.GetHotVideoCache(ctx, version, cursor, limit, &result); err == nil && ok {
		return &result, nil
	}

	videoItems, err := s.db.ListHotVideos(ctx, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	result, err = buildHotVideoListResult(videoItems, limit)
	if err != nil {
		return nil, err
	}
	_ = s.cache.SetHotVideoCache(ctx, version, cursor, limit, result)
	return &result, nil
}

func buildVideoListResult(videos []model.Video, cursor uint, limit int) *VideoListResult {
	hasMore := len(videos) > limit
	if hasMore {
		videos = videos[:limit]
	}

	nextCursor := uint(0)
	if hasMore {
		nextCursor = cursor + uint(len(videos))
	}

	return &VideoListResult{
		Items:      videos,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

func buildHotVideoListResult(videos []model.Video, limit int) (VideoListResult, error) {
	hasMore := len(videos) > limit
	if hasMore {
		videos = videos[:limit]
	}

	nextCursor := ""
	if hasMore && len(videos) > 0 {
		last := videos[len(videos)-1]
		var err error
		nextCursor, err = parser.EncodeHotVideoCursor(parser.HotVideoCursorValue{
			LikeCount:  last.LikeCount,
			VisitCount: last.VisitCount,
			ID:         last.ID,
		})
		if err != nil {
			return VideoListResult{}, err
		}
	}

	return VideoListResult{
		Items:           videos,
		NextCursorToken: nextCursor,
		HasMore:         hasMore,
	}, nil
}
