package service

import "errors"

var (
	ErrUserExists           = errors.New("用户名已存在")
	ErrUserNotFound         = errors.New("用户不存在")
	ErrPasswordWrong        = errors.New("密码错误")
	ErrTokenExpired         = errors.New("令牌已过期")
	ErrTokenInvalid         = errors.New("令牌无效")
	ErrUnsupportedAvatarExt = errors.New("不支持的头像文件类型")
)
