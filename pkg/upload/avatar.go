package upload

import (
	"errors"
)

const (
	AvatarDir   = "./storage/avatars"
	AvatarRoute = "/static/avatars"
)

var ErrUnsupportedAvatarExt = errors.New("不支持的头像文件类型")

var avatarExts = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
}

func PrepareAvatar(userID uint, originalFilename string) (savePath, avatarURL string, err error) {
	return prepareUploadedFile(userID, originalFilename, AvatarDir, AvatarRoute, "user", avatarExts, ErrUnsupportedAvatarExt)
}

func RemoveAvatar(avatarURL string) error {
	return removeUploadedFile(avatarURL, AvatarRoute, AvatarDir)
}
