package response

import (
	v1 "video-platform/biz/model/platform"
)

// 业务错误码定义
const (
	CodeSuccess               = 200
	CodeParamError            = 400
	CodeUnauthorized          = 401
	CodeForbidden             = 403
	CodeNotFound              = 404
	CodeUserExists            = 1001
	CodeUserNotFound          = 1002
	CodePasswordWrong         = 1003
	CodeTokenExpired          = 1004
	CodeTokenInvalid          = 1005
	CodeUnsupportedAvatarType = 1006
	CodeCannotFollowSelf      = 1007
	CodeFollowNotFound        = 1008
	CodeAlreadyFollowed       = 1009
	CodeVideoNotFound         = 1010
	CodeVideoTitleRequired    = 1011
	CodeVideoFileRequired     = 1012
	CodeUnsupportedVideoType  = 1013
	CodeUnsupportedCoverType  = 1014
	CodeInternalError         = 500
)

// 错误信息映射
var codeMsg = map[int32]string{
	CodeSuccess:               "成功",
	CodeParamError:            "参数错误",
	CodeUnauthorized:          "未授权",
	CodeForbidden:             "禁止访问",
	CodeNotFound:              "资源不存在",
	CodeUserExists:            "用户名已存在",
	CodeUserNotFound:          "用户不存在",
	CodePasswordWrong:         "密码错误",
	CodeTokenExpired:          "令牌已过期",
	CodeTokenInvalid:          "令牌无效",
	CodeUnsupportedAvatarType: "不支持的头像文件类型",
	CodeCannotFollowSelf:      "不能关注自己",
	CodeFollowNotFound:        "关注关系不存在",
	CodeAlreadyFollowed:       "已经关注该用户",
	CodeVideoNotFound:         "视频不存在",
	CodeVideoTitleRequired:    "视频标题不能为空",
	CodeVideoFileRequired:     "视频文件不能为空",
	CodeUnsupportedVideoType:  "不支持的视频文件类型",
	CodeUnsupportedCoverType:  "不支持的视频封面文件类型",
	CodeInternalError:         "服务器内部错误",
}

// Success 创建成功响应
func Success(msg ...string) *v1.BaseResponse {
	message := "成功"
	if len(msg) > 0 {
		message = msg[0]
	}
	return &v1.BaseResponse{
		Code: CodeSuccess,
		Msg:  message,
	}
}

// Error 创建错误响应
func Error(code int32, msg ...string) *v1.BaseResponse {
	message := codeMsg[code]
	if len(msg) > 0 {
		message = msg[0]
	}
	if message == "" {
		message = "未知错误"
	}
	return &v1.BaseResponse{
		Code: code,
		Msg:  message,
	}
}

// ParamError 参数错误
func ParamError(msg ...string) *v1.BaseResponse {
	return Error(CodeParamError, msg...)
}

// Unauthorized 未授权
func Unauthorized(msg ...string) *v1.BaseResponse {
	return Error(CodeUnauthorized, msg...)
}

// Forbidden 禁止访问
func Forbidden(msg ...string) *v1.BaseResponse {
	return Error(CodeForbidden, msg...)
}

// NotFound 资源不存在
func NotFound(msg ...string) *v1.BaseResponse {
	return Error(CodeNotFound, msg...)
}

// InternalError 服务器内部错误
func InternalError(msg ...string) *v1.BaseResponse {
	return Error(CodeInternalError, msg...)
}
