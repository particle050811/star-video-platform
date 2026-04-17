package upload

import "errors"

var ErrUnsupportedVideoCoverExt = errors.New("不支持的视频封面文件类型")

var videoCoverExts = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
}

func PrepareVideoCover(userID uint, originalFilename string) (savePath, coverURL string, err error) {
	return prepareUploadedFile(userID, originalFilename, VideoCoverDir, VideoCoverURL, "video_cover", videoCoverExts, ErrUnsupportedVideoCoverExt)
}

func RemoveVideoCover(coverURL string) error {
	return removeUploadedFile(coverURL, VideoCoverURL, VideoCoverDir)
}
