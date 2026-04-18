package parser

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
)

type HotVideoCursorValue struct {
	LikeCount  int64 `json:"like_count"`
	VisitCount int64 `json:"visit_count"`
	ID         uint  `json:"id"`
}

func UserID(rawUserID string) (uint, error) {
	if rawUserID == "" {
		return 0, errors.New("user_id 不能为空")
	}

	parsedUserID, err := strconv.ParseUint(rawUserID, 10, 64)
	if err != nil || parsedUserID == 0 {
		return 0, errors.New("user_id 格式错误")
	}

	return uint(parsedUserID), nil
}

func VideoID(rawVideoID string) (uint, error) {
	if rawVideoID == "" {
		return 0, errors.New("video_id 不能为空")
	}

	parsedVideoID, err := strconv.ParseUint(rawVideoID, 10, 64)
	if err != nil || parsedVideoID == 0 {
		return 0, errors.New("video_id 格式错误")
	}

	return uint(parsedVideoID), nil
}

func ChatRoomID(rawRoomID string) (uint, error) {
	if rawRoomID == "" {
		return 0, errors.New("room_id 不能为空")
	}

	parsedRoomID, err := strconv.ParseUint(rawRoomID, 10, 64)
	if err != nil || parsedRoomID == 0 {
		return 0, errors.New("room_id 格式错误")
	}

	return uint(parsedRoomID), nil
}

func ChatMessageID(rawMessageID string) (uint, error) {
	if rawMessageID == "" {
		return 0, errors.New("message_id 不能为空")
	}

	parsedMessageID, err := strconv.ParseUint(rawMessageID, 10, 64)
	if err != nil || parsedMessageID == 0 {
		return 0, errors.New("message_id 格式错误")
	}

	return uint(parsedMessageID), nil
}

func Cursor(rawCursor string) (uint, error) {
	if rawCursor == "" {
		return 0, nil
	}

	parsedCursor, err := strconv.ParseUint(rawCursor, 10, 64)
	if err != nil {
		return 0, errors.New("cursor 格式错误")
	}

	return uint(parsedCursor), nil
}

func ParseHotVideoCursor(rawCursor string) (HotVideoCursorValue, error) {
	if rawCursor == "" {
		return HotVideoCursorValue{}, nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(rawCursor)
	if err != nil {
		return HotVideoCursorValue{}, errors.New("cursor 格式错误")
	}

	var cursor HotVideoCursorValue
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return HotVideoCursorValue{}, errors.New("cursor 格式错误")
	}
	if cursor.ID == 0 || cursor.LikeCount < 0 || cursor.VisitCount < 0 {
		return HotVideoCursorValue{}, errors.New("cursor 格式错误")
	}

	return cursor, nil
}

func EncodeHotVideoCursor(cursor HotVideoCursorValue) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}
