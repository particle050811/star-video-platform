package repository

import (
	"context"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	"video-platform/biz/dal/rdb"
)

func CreateVideo(ctx context.Context, video *model.Video) error {
	if err := db.CreateVideo(ctx, video); err != nil {
		return err
	}

	_ = rdb.BumpHotVideoCacheVersion(ctx)
	return nil
}

func GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	var video model.Video
	if ok, err := rdb.GetVideoDetailCache(ctx, videoID, &video); err == nil && ok {
		return &video, nil
	}

	fetched, err := db.GetVideoByID(ctx, videoID)
	if err != nil {
		return nil, err
	}

	_ = rdb.SetVideoDetailCache(ctx, videoID, fetched)
	return fetched, nil
}

func ListVideosByUserID(ctx context.Context, userID uint, offset, limit int) ([]model.Video, error) {
	return db.ListVideosByUserID(ctx, userID, offset, limit)
}

func SearchVideos(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, offset, limit int) ([]model.Video, error) {
	return db.SearchVideos(ctx, db.VideoQuery{
		Keywords: keywords,
		UserIDs:  userIDs,
		FromDate: fromDate,
		ToDate:   toDate,
		SortBy:   sortBy,
		Offset:   offset,
		Limit:    limit,
	})
}

func ListHotVideos(ctx context.Context, offset, limit int) ([]model.Video, error) {
	version, err := rdb.GetHotVideoCacheVersion(ctx)
	if err != nil {
		return nil, err
	}

	var videos []model.Video
	if ok, err := rdb.GetHotVideoCache(ctx, version, offset, limit, &videos); err == nil && ok {
		return videos, nil
	}

	videos, err = db.ListHotVideos(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	_ = rdb.SetHotVideoCache(ctx, version, offset, limit, videos)
	return videos, nil
}
