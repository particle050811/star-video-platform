package repository

import (
	"context"
	dbdal "video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	rdbdal "video-platform/biz/dal/rdb"
)

type videoDBStore interface {
	CreateVideo(ctx context.Context, video *model.Video) error
	GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error)
	ListVideosByUserID(ctx context.Context, userID uint, offset, limit int) ([]model.Video, error)
	SearchVideos(ctx context.Context, params dbdal.VideoQuery) ([]model.Video, error)
	ListHotVideos(ctx context.Context, offset, limit int) ([]model.Video, error)
}

type videoCacheStore interface {
	BumpHotVideoCacheVersion(ctx context.Context) error
	GetVideoDetailCache(ctx context.Context, videoID uint, dest any) (bool, error)
	SetVideoDetailCache(ctx context.Context, videoID uint, value any) error
	GetHotVideoCacheVersion(ctx context.Context) (int64, error)
	GetHotVideoCache(ctx context.Context, version int64, offset, limit int, dest any) (bool, error)
	SetHotVideoCache(ctx context.Context, version int64, offset, limit int, value any) error
}

type videoStore struct {
	db    videoDBStore
	cache videoCacheStore
}

var videos = videoStore{
	db:    dbdal.Videos,
	cache: rdbdal.Videos,
}

func CreateVideo(ctx context.Context, video *model.Video) error {
	return videos.CreateVideo(ctx, video)
}

func GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return videos.GetVideoByID(ctx, videoID)
}

func ListVideosByUserID(ctx context.Context, userID uint, offset, limit int) ([]model.Video, error) {
	return videos.ListVideosByUserID(ctx, userID, offset, limit)
}

func SearchVideos(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, offset, limit int) ([]model.Video, error) {
	return videos.SearchVideos(ctx, keywords, userIDs, fromDate, toDate, sortBy, offset, limit)
}

func ListHotVideos(ctx context.Context, offset, limit int) ([]model.Video, error) {
	return videos.ListHotVideos(ctx, offset, limit)
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

func (s videoStore) ListVideosByUserID(ctx context.Context, userID uint, offset, limit int) ([]model.Video, error) {
	return s.db.ListVideosByUserID(ctx, userID, offset, limit)
}

func (s videoStore) SearchVideos(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, offset, limit int) ([]model.Video, error) {
	return s.db.SearchVideos(ctx, dbdal.VideoQuery{
		Keywords: keywords,
		UserIDs:  userIDs,
		FromDate: fromDate,
		ToDate:   toDate,
		SortBy:   sortBy,
		Offset:   offset,
		Limit:    limit,
	})
}

func (s videoStore) ListHotVideos(ctx context.Context, offset, limit int) ([]model.Video, error) {
	version, err := s.cache.GetHotVideoCacheVersion(ctx)
	if err != nil {
		return nil, err
	}

	var videos []model.Video
	if ok, err := s.cache.GetHotVideoCache(ctx, version, offset, limit, &videos); err == nil && ok {
		return videos, nil
	}

	videos, err = s.db.ListHotVideos(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetHotVideoCache(ctx, version, offset, limit, videos)
	return videos, nil
}
