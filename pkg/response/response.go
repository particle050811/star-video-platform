package response

import (
	v1 "video-platform/biz/model/platform"
)

// 业务错误码定义
const (
	CodeSuccess = 10000 + iota
	CodeParamError
	CodeUnauthorized
	CodeForbidden
	CodeNotFound
	CodeUserExists
	CodeUserNotFound
	CodePasswordWrong
	CodeTokenExpired
	CodeTokenInvalid
	CodeInternalError
)

// 错误信息映射
var codeMsg = map[int32]string{
	CodeSuccess:       "成功",
	CodeParamError:    "参数错误",
	CodeUnauthorized:  "未授权",
	CodeForbidden:     "禁止访问",
	CodeNotFound:      "资源不存在",
	CodeUserExists:    "用户名已存在",
	CodeUserNotFound:  "用户不存在",
	CodePasswordWrong: "密码错误",
	CodeTokenExpired:  "令牌已过期",
	CodeTokenInvalid:  "令牌无效",
	CodeInternalError: "服务器内部错误",
}

// Success 创建成功响应
func Success() *v1.BaseResponse {
	return &v1.BaseResponse{
		Code: CodeSuccess,
		Msg:  codeMsg[CodeSuccess],
	}
}

// Error 创建错误响应
func Error(code int32) *v1.BaseResponse {
	return &v1.BaseResponse{
		Code: code,
		Msg:  codeMsg[code],
	}
}

// ParamError 参数错误
func ParamError() *v1.BaseResponse {
	return Error(CodeParamError)
}

// Unauthorized 未授权
func Unauthorized() *v1.BaseResponse {
	return Error(CodeUnauthorized)
}

// Forbidden 禁止访问
func Forbidden() *v1.BaseResponse {
	return Error(CodeForbidden)
}

// NotFound 资源不存在
func NotFound() *v1.BaseResponse {
	return Error(CodeNotFound)
}

// InternalError 服务器内部错误
func InternalError() *v1.BaseResponse {
	return Error(CodeInternalError)
}
