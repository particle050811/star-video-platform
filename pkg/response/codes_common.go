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

var commonCodeMsg = map[int32]string{
	CodeSuccess:       "成功",
	CodeParamError:    "参数错误",
	CodeUnauthorized:  "未授权",
	CodeForbidden:     "禁止访问",
	CodeNotFound:      "资源不存在",
	CodeInternalError: "服务器内部错误",
}
