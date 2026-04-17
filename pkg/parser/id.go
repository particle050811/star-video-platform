package parser

import (
	"errors"
	"strconv"
)

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
