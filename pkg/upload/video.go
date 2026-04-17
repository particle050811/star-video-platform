package upload

import (
	"errors"
)

const (
	VideoDir      = "./storage/videos"
	VideoRoute    = "/static/videos"
	VideoCoverDir = "./storage/video-covers"
	VideoCoverURL = "/static/video-covers"
)

var (
	ErrUnsupportedVideoExt = errors.New("不支持的视频文件类型")
)

var videoExts = map[string]struct{}{
	".mp4":  {},
	".mov":  {},
	".m4v":  {},
	".webm": {},
}

func PrepareVideo(userID uint, originalFilename string) (savePath, videoURL string, err error) {
	return prepareUploadedFile(userID, originalFilename, VideoDir, VideoRoute, "video", videoExts, ErrUnsupportedVideoExt)
}

func RemoveVideo(videoURL string) error {
	return removeUploadedFile(videoURL, VideoRoute, VideoDir)
}
