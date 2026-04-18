package response

// 通用响应码
const (
	CodeSuccess       = 200
	CodeParamError    = 400
	CodeUnauthorized  = 401
	CodeForbidden     = 403
	CodeNotFound      = 404
	CodeInternalError = 500
)

// 用户模块响应码
const (
	CodeUserExists            = 1001
	CodeUserNotFound          = 1002
	CodePasswordWrong         = 1003
	CodeTokenExpired          = 1004
	CodeTokenInvalid          = 1005
	CodeUnsupportedAvatarType = 1006
)

// 关系模块响应码
const (
	CodeCannotFollowSelf = 1007
	CodeFollowNotFound   = 1008
	CodeAlreadyFollowed  = 1009
)

// 视频模块响应码
const (
	CodeVideoNotFound        = 1010
	CodeVideoTitleRequired   = 1011
	CodeVideoFileRequired    = 1012
	CodeUnsupportedVideoType = 1013
	CodeUnsupportedCoverType = 1014
)

// 互动模块响应码
const (
	CodeCommentNotFound = 1015
	CodeNoPermission    = 1016
	CodeCommentTooLong  = 1017
	CodeCommentEmpty    = 1018
	CodeAlreadyLiked    = 1019
	CodeLikeNotFound    = 1020
)

// 聊天模块响应码
const (
	CodeChatRoomNotFound       = 1021
	CodeChatRoomMemberNotFound = 1022
	CodeChatMemberExists       = 1023
	CodeChatMemberRequired     = 1024
	CodeChatPrivateMemberCount = 1025
	CodeChatGroupNameRequired  = 1026
	CodeChatOwnerCannotLeave   = 1027
	CodeChatMessageNotFound    = 1028
)
