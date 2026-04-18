package relation

import "errors"

var (
	ErrCannotFollowSelf = errors.New("不能关注自己")
	ErrAlreadyFollowed  = errors.New("已经关注该用户")
	ErrFollowNotFound   = errors.New("关注关系不存在")
)
