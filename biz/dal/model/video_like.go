package model

import "time"

// VideoLike 用户点赞视频关系。
type VideoLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_video_like" json:"user_id"`
	VideoID   uint      `gorm:"not null;uniqueIndex:idx_user_video_like;index:idx_video_likes_video_id" json:"video_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名。
func (VideoLike) TableName() string {
	return "video_likes"
}
