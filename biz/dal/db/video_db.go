package db

import (
	"context"
	"strings"
	"time"
	"video-platform/biz/dal/model"

	"gorm.io/gorm"
)

type VideoQuery struct {
	Keywords string
	UserIDs  []uint
	FromDate int64
	ToDate   int64
	SortBy   string
	Cursor   uint
	Limit    int
}

type VideoDB struct {
	db *gorm.DB
}

func NewVideoDB(gdb *gorm.DB) VideoDB {
	return VideoDB{db: gdb}
}

var Videos = NewVideoDB(DB)

func (v VideoDB) gormDB() *gorm.DB {
	if v.db != nil {
		return v.db
	}
	return DB
}

func (v VideoDB) CreateVideo(ctx context.Context, video *model.Video) error {
	return v.gormDB().WithContext(ctx).Create(video).Error
}

func (v VideoDB) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	var video model.Video
	if err := v.gormDB().WithContext(ctx).First(&video, videoID).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

func (v VideoDB) ListVideosByUserID(ctx context.Context, userID uint, cursor uint, limit int) ([]model.Video, error) {
	query := v.gormDB().WithContext(ctx).Model(&model.Video{}).Where("user_id = ?", userID)

	videos := make([]model.Video, 0)
	if err := query.Order("created_at DESC, id DESC").Offset(int(cursor)).Limit(limit).Find(&videos).Error; err != nil {
		return nil, err
	}

	return videos, nil
}

func (v VideoDB) SearchVideos(ctx context.Context, params VideoQuery) ([]model.Video, error) {
	query := v.gormDB().WithContext(ctx).Model(&model.Video{})

	if keywords := strings.TrimSpace(params.Keywords); keywords != "" {
		like := "%" + keywords + "%"
		query = query.Where("videos.title LIKE ? OR videos.description LIKE ?", like, like)
	}

	if params.UserIDs != nil {
		if len(params.UserIDs) == 0 {
			return []model.Video{}, nil
		}
		query = query.Where("user_id IN ?", params.UserIDs)
	}

	if params.FromDate > 0 {
		query = query.Where("videos.created_at >= ?", time.Unix(params.FromDate, 0))
	}

	if params.ToDate > 0 {
		query = query.Where("videos.created_at <= ?", time.Unix(params.ToDate, 0))
	}

	videos := make([]model.Video, 0)
	orderBy := "videos.created_at DESC, videos.id DESC"
	if strings.EqualFold(params.SortBy, "hot") {
		orderBy = "videos.like_count DESC, videos.visit_count DESC, videos.id DESC"
	}

	if err := query.Order(orderBy).Offset(int(params.Cursor)).Limit(params.Limit).Find(&videos).Error; err != nil {
		return nil, err
	}

	return videos, nil
}

func (v VideoDB) ListHotVideos(ctx context.Context, cursor uint, limit int) ([]model.Video, error) {
	query := v.gormDB().WithContext(ctx).Model(&model.Video{})

	videos := make([]model.Video, 0)
	if err := query.Order("like_count DESC, visit_count DESC, id DESC").Offset(int(cursor)).Limit(limit).Find(&videos).Error; err != nil {
		return nil, err
	}

	return videos, nil
}

func CreateVideo(ctx context.Context, video *model.Video) error {
	return Videos.CreateVideo(ctx, video)
}

func GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return Videos.GetVideoByID(ctx, videoID)
}

func ListVideosByUserID(ctx context.Context, userID uint, cursor uint, limit int) ([]model.Video, error) {
	return Videos.ListVideosByUserID(ctx, userID, cursor, limit)
}

func SearchVideos(ctx context.Context, params VideoQuery) ([]model.Video, error) {
	return Videos.SearchVideos(ctx, params)
}

func ListHotVideos(ctx context.Context, cursor uint, limit int) ([]model.Video, error) {
	return Videos.ListHotVideos(ctx, cursor, limit)
}
