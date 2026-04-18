package response

import (
	v1 "video-platform/biz/model/platform"
)

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
