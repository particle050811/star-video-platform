// Package service 业务逻辑层
package service

import "errors"

var (
	ErrNoPermission = errors.New("无权限操作")
)
