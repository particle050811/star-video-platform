package service

import "errors"

var (
	ErrVideoNotFound            = errors.New("视频不存在")
	ErrVideoTitleRequired       = errors.New("视频标题不能为空")
	ErrVideoFileRequired        = errors.New("视频文件不能为空")
	ErrUnsupportedVideoExt      = errors.New("不支持的视频文件类型")
	ErrUnsupportedVideoCoverExt = errors.New("不支持的视频封面文件类型")
)
