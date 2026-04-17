package repository

import (
	"context"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
)

type VideoQuery struct {
	Keywords string
	UserIDs  []uint
	FromDate int64
	ToDate   int64
	SortBy   string
	Offset   int
	Limit    int
}

func CreateVideo(ctx context.Context, video *model.Video) error {
	return db.CreateVideo(ctx, video)
}

func GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return db.GetVideoByID(ctx, videoID)
}

func ListVideosByUserID(ctx context.Context, userID uint, offset, limit int) ([]model.Video, error) {
	return db.ListVideosByUserID(ctx, userID, offset, limit)
}

func SearchVideos(ctx context.Context, params VideoQuery) ([]model.Video, error) {
	return db.SearchVideos(ctx, db.VideoQuery{
		Keywords: params.Keywords,
		UserIDs:  params.UserIDs,
		FromDate: params.FromDate,
		ToDate:   params.ToDate,
		SortBy:   params.SortBy,
		Offset:   params.Offset,
		Limit:    params.Limit,
	})
}

func ListHotVideos(ctx context.Context, offset, limit int) ([]model.Video, error) {
	return db.ListHotVideos(ctx, offset, limit)
}
