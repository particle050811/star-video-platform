package model

import "time"

// Relation 用户关注关系，from_user_id 关注 to_user_id。
type Relation struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FromUserID uint      `gorm:"not null;uniqueIndex:idx_from_to" json:"from_user_id"`
	ToUserID   uint      `gorm:"not null;uniqueIndex:idx_from_to;index:idx_to_user_id" json:"to_user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名。
func (Relation) TableName() string {
	return "relations"
}
