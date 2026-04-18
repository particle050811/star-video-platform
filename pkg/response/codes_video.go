package response

// 视频模块响应码
const (
	CodeVideoNotFound        = 1010
	CodeVideoTitleRequired   = 1011
	CodeVideoFileRequired    = 1012
	CodeUnsupportedVideoType = 1013
	CodeUnsupportedCoverType = 1014
)

var videoCodeMsg = map[int32]string{
	CodeVideoNotFound:        "视频不存在",
	CodeVideoTitleRequired:   "视频标题不能为空",
	CodeVideoFileRequired:    "视频文件不能为空",
	CodeUnsupportedVideoType: "不支持的视频文件类型",
	CodeUnsupportedCoverType: "不支持的视频封面文件类型",
}
