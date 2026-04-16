package upload

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	AvatarDir   = "./storage/avatars"
	AvatarRoute = "/static/avatars"
)

var ErrUnsupportedAvatarExt = errors.New("不支持的头像文件类型")

func PrepareAvatar(userID uint, originalFilename string) (savePath, avatarURL string, err error) {
	if err := os.MkdirAll(AvatarDir, 0o755); err != nil {
		return "", "", err
	}

	ext := strings.ToLower(filepath.Ext(originalFilename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		return "", "", ErrUnsupportedAvatarExt
	}

	filename := fmt.Sprintf("user_%d_%d%s", userID, time.Now().UnixNano(), ext)
	return filepath.Join(AvatarDir, filename), AvatarRoute + "/" + filename, nil
}
