package model

import (
	"time"

	"gorm.io/gorm"
)

// ChatRoom 统一表示私聊和群聊房间。
type ChatRoom struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Type          string         `gorm:"size:20;not null;index:idx_chat_rooms_type" json:"type"`
	Name          string         `gorm:"size:100" json:"name"`
	OwnerID       uint           `gorm:"index:idx_chat_rooms_owner_id" json:"owner_id"`
	LastMessageID uint           `gorm:"index:idx_chat_rooms_last_message_id" json:"last_message_id"`
	LastMessageAt *time.Time     `gorm:"index:idx_chat_rooms_last_message_at" json:"last_message_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (ChatRoom) TableName() string {
	return "chat_rooms"
}

// ChatRoomMember 表示用户与房间的成员关系。
type ChatRoomMember struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	RoomID        uint      `gorm:"not null;uniqueIndex:idx_chat_room_member;index:idx_chat_room_members_room_id" json:"room_id"`
	UserID        uint      `gorm:"not null;uniqueIndex:idx_chat_room_member;index:idx_chat_room_members_user_id" json:"user_id"`
	Role          string    `gorm:"size:20;not null" json:"role"`
	LastReadMsgID uint      `json:"last_read_msg_id"`
	JoinedAt      time.Time `gorm:"index:idx_chat_room_members_joined_at" json:"joined_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (ChatRoomMember) TableName() string {
	return "chat_room_members"
}

// ChatMessage 表示房间内的一条消息。
type ChatMessage struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	RoomID      uint           `gorm:"not null;index:idx_chat_messages_room_id_created_at;index:idx_chat_messages_room_id_id" json:"room_id"`
	SenderID    uint           `gorm:"not null;index:idx_chat_messages_sender_id" json:"sender_id"`
	MessageType string         `gorm:"size:20;not null" json:"message_type"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Status      string         `gorm:"size:20;not null" json:"status"`
	ClientMsgID string         `gorm:"size:100;index:idx_chat_messages_client_msg_id" json:"client_msg_id"`
	CreatedAt   time.Time      `gorm:"index:idx_chat_messages_room_id_created_at" json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
