package db

import (
	"context"
	"errors"
	"strings"
	"time"
	"video-platform/biz/dal/model"
	"video-platform/pkg/parser"

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

var ErrVideoCursorInvalid = errors.New("video cursor is invalid")

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

func (v VideoDB) ListVideosByIDs(ctx context.Context, videoIDs []uint) ([]model.Video, error) {
	videos := make([]model.Video, 0, len(videoIDs))
	if len(videoIDs) == 0 {
		return videos, nil
	}

	if err := v.gormDB().WithContext(ctx).
		Where("id IN ?", videoIDs).
		Find(&videos).Error; err != nil {
		return nil, err
	}

	videoMap := make(map[uint]model.Video, len(videos))
	for _, video := range videos {
		videoMap[video.ID] = video
	}

	ordered := make([]model.Video, 0, len(videoIDs))
	for _, videoID := range videoIDs {
		video, ok := videoMap[videoID]
		if !ok {
			continue
		}
		ordered = append(ordered, video)
	}

	return ordered, nil
}

func (v VideoDB) ListVideosByUserID(ctx context.Context, userID uint, cursor uint, limit int) ([]model.Video, error) {
	query := v.gormDB().WithContext(ctx).Model(&model.Video{}).Where("user_id = ?", userID)
	if cursor > 0 {
		anchor, err := v.getVideoCursorAnchor(ctx, cursor, func(query *gorm.DB) *gorm.DB {
			return query.Where("user_id = ?", userID)
		})
		if err != nil {
			return nil, err
		}
		if anchor == nil {
			return nil, ErrVideoCursorInvalid
		}
		query = query.Where(
			"(created_at < ?) OR (created_at = ? AND id < ?)",
			anchor.CreatedAt,
			anchor.CreatedAt,
			anchor.ID,
		)
	}

	videos := make([]model.Video, 0)
	if err := query.Order("created_at DESC, id DESC").Limit(limit).Find(&videos).Error; err != nil {
		return nil, err
	}

	return videos, nil
}

func (v VideoDB) SearchVideos(ctx context.Context, params VideoQuery) ([]model.Video, error) {
	query := v.gormDB().WithContext(ctx).Model(&model.Video{})

	query = applyVideoSearchFilters(query, params)

	if params.UserIDs != nil {
		if len(params.UserIDs) == 0 {
			return []model.Video{}, nil
		}
	}

	videos := make([]model.Video, 0)
	orderBy := "videos.created_at DESC, videos.id DESC"
	if params.Cursor > 0 {
		anchor, err := v.getVideoCursorAnchor(ctx, params.Cursor, func(query *gorm.DB) *gorm.DB {
			return applyVideoSearchFilters(query, params)
		})
		if err != nil {
			return nil, err
		}

		if anchor == nil {
			return nil, ErrVideoCursorInvalid
		}
		if strings.EqualFold(params.SortBy, "hot") {
			query = query.Where(
				"(videos.like_count < ?) OR (videos.like_count = ? AND videos.visit_count < ?) OR (videos.like_count = ? AND videos.visit_count = ? AND videos.id < ?)",
				anchor.LikeCount,
				anchor.LikeCount,
				anchor.VisitCount,
				anchor.LikeCount,
				anchor.VisitCount,
				anchor.ID,
			)
		} else {
			query = query.Where(
				"(videos.created_at < ?) OR (videos.created_at = ? AND videos.id < ?)",
				anchor.CreatedAt,
				anchor.CreatedAt,
				anchor.ID,
			)
		}
	}
	if strings.EqualFold(params.SortBy, "hot") {
		orderBy = "videos.like_count DESC, videos.visit_count DESC, videos.id DESC"
	}

	if err := query.Order(orderBy).Limit(params.Limit).Find(&videos).Error; err != nil {
		return nil, err
	}

	return videos, nil
}

func applyVideoSearchFilters(query *gorm.DB, params VideoQuery) *gorm.DB {
	if keywords := strings.TrimSpace(params.Keywords); keywords != "" {
		like := "%" + escapeLikePattern(keywords) + "%"
		query = query.Where("videos.title LIKE ? ESCAPE '\\\\' OR videos.description LIKE ? ESCAPE '\\\\'", like, like)
	}

	if params.UserIDs != nil && len(params.UserIDs) > 0 {
		query = query.Where("user_id IN ?", params.UserIDs)
	}

	if params.FromDate > 0 {
		query = query.Where("videos.created_at >= ?", time.Unix(params.FromDate, 0))
	}

	if params.ToDate > 0 {
		query = query.Where("videos.created_at <= ?", time.Unix(params.ToDate, 0))
	}

	return query
}

func (v VideoDB) ListHotVideos(ctx context.Context, cursor parser.HotVideoCursorValue, limit int) ([]model.Video, error) {
	query := v.gormDB().WithContext(ctx).Model(&model.Video{})
	if cursor.ID > 0 {
		query = query.Where(
			"like_count < ? OR (like_count = ? AND visit_count < ?) OR (like_count = ? AND visit_count = ? AND id < ?)",
			cursor.LikeCount,
			cursor.LikeCount,
			cursor.VisitCount,
			cursor.LikeCount,
			cursor.VisitCount,
			cursor.ID,
		)
	}

	videos := make([]model.Video, 0)
	if err := query.Order("like_count DESC, visit_count DESC, id DESC").Limit(limit).Find(&videos).Error; err != nil {
		return nil, err
	}

	return videos, nil
}

func (v VideoDB) getVideoCursorAnchor(ctx context.Context, cursor uint, scope func(*gorm.DB) *gorm.DB) (*model.Video, error) {
	var anchor model.Video
	query := v.gormDB().WithContext(ctx).
		Select("id", "created_at", "like_count", "visit_count").
		Where("id = ?", cursor)
	if scope != nil {
		query = scope(query)
	}
	if err := query.First(&anchor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &anchor, nil
}

func CreateVideo(ctx context.Context, video *model.Video) error {
	return Videos.CreateVideo(ctx, video)
}

func GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return Videos.GetVideoByID(ctx, videoID)
}

func ListVideosByIDs(ctx context.Context, videoIDs []uint) ([]model.Video, error) {
	return Videos.ListVideosByIDs(ctx, videoIDs)
}

func ListVideosByUserID(ctx context.Context, userID uint, cursor uint, limit int) ([]model.Video, error) {
	return Videos.ListVideosByUserID(ctx, userID, cursor, limit)
}

func SearchVideos(ctx context.Context, params VideoQuery) ([]model.Video, error) {
	return Videos.SearchVideos(ctx, params)
}

func ListHotVideos(ctx context.Context, cursor parser.HotVideoCursorValue, limit int) ([]model.Video, error) {
	return Videos.ListHotVideos(ctx, cursor, limit)
}
