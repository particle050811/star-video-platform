package upload

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
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

func SaveFile(file *multipart.FileHeader, savePath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(savePath), 0o755); err != nil {
		return err
	}

	dst, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func RemoveAvatar(avatarURL string) error {
	if avatarURL == "" {
		return nil
	}

	filename := strings.TrimPrefix(avatarURL, AvatarRoute+"/")
	err := os.Remove(filepath.Join(AvatarDir, filename))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
