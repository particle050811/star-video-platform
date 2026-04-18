package response

// 关系模块响应码
const (
	CodeCannotFollowSelf = 1007
	CodeFollowNotFound   = 1008
	CodeAlreadyFollowed  = 1009
)

var relationCodeMsg = map[int32]string{
	CodeCannotFollowSelf: "不能关注自己",
	CodeFollowNotFound:   "关注关系不存在",
	CodeAlreadyFollowed:  "已经关注该用户",
}
