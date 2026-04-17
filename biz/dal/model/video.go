package model

import (
	"time"

	"gorm.io/gorm"
)

// Video 视频模型
type Video struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index:idx_videos_user_id" json:"user_id"`
	VideoURL     string         `gorm:"size:500;not null" json:"video_url"`
	CoverURL     string         `gorm:"size:500" json:"cover_url"`
	Title        string         `gorm:"size:255;not null" json:"title"`
	Description  string         `gorm:"type:text" json:"description"`
	VisitCount   int64          `gorm:"default:0;index:idx_videos_visit_count" json:"visit_count"`
	LikeCount    int64          `gorm:"default:0;index:idx_videos_like_count" json:"like_count"`
	CommentCount int64          `gorm:"default:0;index:idx_videos_comment_count" json:"comment_count"`
	CreatedAt    time.Time      `gorm:"index:idx_videos_created_at" json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (Video) TableName() string {
	return "videos"
}
