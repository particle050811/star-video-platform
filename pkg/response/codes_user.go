package response

// 用户模块响应码
const (
	CodeUserExists            = 1001
	CodeUserNotFound          = 1002
	CodePasswordWrong         = 1003
	CodeTokenExpired          = 1004
	CodeTokenInvalid          = 1005
	CodeUnsupportedAvatarType = 1006
)

var userCodeMsg = map[int32]string{
	CodeUserExists:            "用户名已存在",
	CodeUserNotFound:          "用户不存在",
	CodePasswordWrong:         "密码错误",
	CodeTokenExpired:          "令牌已过期",
	CodeTokenInvalid:          "令牌无效",
	CodeUnsupportedAvatarType: "不支持的头像文件类型",
}
