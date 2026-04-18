package response

// 互动模块响应码
const (
	CodeCommentNotFound = 1015
	CodeNoPermission    = 1016
	CodeCommentTooLong  = 1017
	CodeCommentEmpty    = 1018
	CodeAlreadyLiked    = 1019
	CodeLikeNotFound    = 1020
)

var interactionCodeMsg = map[int32]string{
	CodeCommentNotFound: "评论不存在",
	CodeNoPermission:    "无权限操作",
	CodeCommentTooLong:  "评论内容过长",
	CodeCommentEmpty:    "评论内容不能为空",
	CodeAlreadyLiked:    "已经点赞该视频",
	CodeLikeNotFound:    "点赞记录不存在",
}
