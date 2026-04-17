package repository

import (
	"context"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
)

func CreateVideo(ctx context.Context, video *model.Video) error {
	return db.CreateVideo(ctx, video)
}

func GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return db.GetVideoByID(ctx, videoID)
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
	return db.ListHotVideos(ctx, offset, limit)
}
