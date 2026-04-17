package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment 视频评论模型
type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	VideoID   uint           `gorm:"not null;index:idx_comments_video_id" json:"video_id"`
	UserID    uint           `gorm:"not null;index:idx_comments_user_id" json:"user_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	LikeCount int64          `gorm:"default:0;index:idx_comments_like_count" json:"like_count"`
	CreatedAt time.Time      `gorm:"index:idx_comments_created_at" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}
