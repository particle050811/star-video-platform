package db

import (
	"context"
	"time"
	"video-platform/biz/dal/model"

	"gorm.io/gorm"
)

type ChatDB struct {
	db *gorm.DB
}

type ChatRoomListItem struct {
	Room        model.ChatRoom
	UnreadCount int64
	MemberCount int64
}

func NewChatDB(gdb *gorm.DB) ChatDB {
	return ChatDB{db: gdb}
}

var Chats = NewChatDB(DB)

func (c ChatDB) gormDB() *gorm.DB {
	if c.db != nil {
		return c.db
	}
	return DB
}

func (c ChatDB) CreateRoom(ctx context.Context, room *model.ChatRoom, members []model.ChatRoomMember) error {
	return c.gormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(room).Error; err != nil {
			return err
		}

		now := time.Now()
		for i := range members {
			members[i].RoomID = room.ID
			if members[i].JoinedAt.IsZero() {
				members[i].JoinedAt = now
			}
		}
		return tx.Create(&members).Error
	})
}

func (c ChatDB) ListRoomsByUserID(ctx context.Context, userID uint, cursor uint, limit int) ([]ChatRoomListItem, error) {
	members := make([]model.ChatRoomMember, 0)
	memberQuery := c.gormDB().WithContext(ctx).
		Model(&model.ChatRoomMember{}).
		Where("user_id = ?", userID)
	if cursor > 0 {
		memberQuery = memberQuery.Where("room_id < ?", cursor)
	}
	if err := memberQuery.
		Order("room_id DESC").
		Limit(limit).
		Find(&members).Error; err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []ChatRoomListItem{}, nil
	}

	roomIDs := make([]uint, 0, len(members))
	lastReadMsgIDs := make(map[uint]uint, len(members))
	for _, member := range members {
		roomIDs = append(roomIDs, member.RoomID)
		lastReadMsgIDs[member.RoomID] = member.LastReadMsgID
	}

	rooms := make([]model.ChatRoom, 0, len(roomIDs))
	if err := c.gormDB().WithContext(ctx).
		Where("id IN ?", roomIDs).
		Find(&rooms).Error; err != nil {
		return nil, err
	}

	roomMap := make(map[uint]model.ChatRoom, len(rooms))
	for _, room := range rooms {
		roomMap[room.ID] = room
	}

	items := make([]ChatRoomListItem, 0, len(members))
	for _, member := range members {
		room, ok := roomMap[member.RoomID]
		if !ok {
			continue
		}

		var unreadCount int64
		if err := c.gormDB().WithContext(ctx).Model(&model.ChatMessage{}).
			Where("room_id = ? AND id > ? AND sender_id <> ?", room.ID, lastReadMsgIDs[room.ID], userID).
			Count(&unreadCount).Error; err != nil {
			return nil, err
		}

		var memberCount int64
		if err := c.gormDB().WithContext(ctx).Model(&model.ChatRoomMember{}).
			Where("room_id = ?", room.ID).
			Count(&memberCount).Error; err != nil {
			return nil, err
		}

		items = append(items, ChatRoomListItem{
			Room:        room,
			UnreadCount: unreadCount,
			MemberCount: memberCount,
		})
	}

	return items, nil
}

func (c ChatDB) GetRoomByID(ctx context.Context, roomID uint) (*model.ChatRoom, error) {
	var room model.ChatRoom
	if err := c.gormDB().WithContext(ctx).First(&room, roomID).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (c ChatDB) GetRoomMember(ctx context.Context, roomID, userID uint) (*model.ChatRoomMember, error) {
	var member model.ChatRoomMember
	if err := c.gormDB().WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (c ChatDB) ListRoomMembers(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatRoomMember, error) {
	members := make([]model.ChatRoomMember, 0)
	query := c.gormDB().WithContext(ctx).Where("room_id = ?", roomID)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	if err := query.Order("id DESC").Limit(limit).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (c ChatDB) ListRoomMessages(ctx context.Context, roomID uint, cursor uint, limit int) ([]model.ChatMessage, error) {
	messages := make([]model.ChatMessage, 0)
	query := c.gormDB().WithContext(ctx).Where("room_id = ?", roomID)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	if err := query.Order("id DESC").Limit(limit).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (c ChatDB) ListMessagesByIDs(ctx context.Context, messageIDs []uint) ([]model.ChatMessage, error) {
	messages := make([]model.ChatMessage, 0, len(messageIDs))
	if len(messageIDs) == 0 {
		return messages, nil
	}
	if err := c.gormDB().WithContext(ctx).Where("id IN ?", messageIDs).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (c ChatDB) AddRoomMembers(ctx context.Context, members []model.ChatRoomMember) error {
	if len(members) == 0 {
		return nil
	}
	return c.gormDB().WithContext(ctx).Create(&members).Error
}

func (c ChatDB) DeleteRoomMember(ctx context.Context, roomID, userID uint) (bool, error) {
	res := c.gormDB().WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Delete(&model.ChatRoomMember{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (c ChatDB) UpdateLastReadMessageID(ctx context.Context, roomID, userID, messageID uint) error {
	if messageID > 0 {
		var count int64
		if err := c.gormDB().WithContext(ctx).Model(&model.ChatMessage{}).
			Where("id = ? AND room_id = ?", messageID, roomID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}

	var count int64
	if err := c.gormDB().WithContext(ctx).Model(&model.ChatRoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}

	return c.gormDB().WithContext(ctx).
		Model(&model.ChatRoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Update("last_read_msg_id", messageID).Error
}

func (c ChatDB) CreateMessage(ctx context.Context, message *model.ChatMessage) error {
	return c.gormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		return tx.Model(&model.ChatRoom{}).
			Where("id = ?", message.RoomID).
			Updates(map[string]any{
				"last_message_id": message.ID,
				"last_message_at": message.CreatedAt,
			}).Error
	})
}
