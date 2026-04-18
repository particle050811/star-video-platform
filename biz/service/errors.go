// Package service 业务逻辑层
package service

import "errors"

// 用户模块错误
var (
	ErrUserExists           = errors.New("用户名已存在")
	ErrUserNotFound         = errors.New("用户不存在")
	ErrPasswordWrong        = errors.New("密码错误")
	ErrTokenExpired         = errors.New("令牌已过期")
	ErrTokenInvalid         = errors.New("令牌无效")
	ErrUnsupportedAvatarExt = errors.New("不支持的头像文件类型")
)

// 视频模块错误
var (
	ErrVideoNotFound            = errors.New("视频不存在")
	ErrVideoTitleRequired       = errors.New("视频标题不能为空")
	ErrVideoFileRequired        = errors.New("视频文件不能为空")
	ErrUnsupportedVideoExt      = errors.New("不支持的视频文件类型")
	ErrUnsupportedVideoCoverExt = errors.New("不支持的视频封面文件类型")
)

// 互动模块错误
var (
	ErrCommentNotFound = errors.New("评论不存在")
	ErrNoPermission    = errors.New("无权限操作")
	ErrCommentTooLong  = errors.New("评论内容过长")
	ErrCommentEmpty    = errors.New("评论内容不能为空")
	ErrAlreadyLiked    = errors.New("已经点赞该视频")
	ErrLikeNotFound    = errors.New("点赞记录不存在")
)

// 社交模块错误
var (
	ErrCannotFollowSelf = errors.New("不能关注自己")
	ErrAlreadyFollowed  = errors.New("已经关注该用户")
	ErrFollowNotFound   = errors.New("关注关系不存在")
)

// 聊天模块错误
var (
	ErrChatRoomNotFound       = errors.New("聊天房间不存在")
	ErrChatRoomMemberNotFound = errors.New("不是聊天房间成员")
	ErrChatMemberExists       = errors.New("用户已在聊天房间中")
	ErrChatMemberRequired     = errors.New("聊天房间成员不能为空")
	ErrChatPrivateMemberCount = errors.New("私聊房间只能包含两个成员")
	ErrChatGroupNameRequired  = errors.New("群聊名称不能为空")
	ErrChatOwnerCannotLeave   = errors.New("群主不能直接退出群聊")
	ErrChatMessageNotFound    = errors.New("聊天消息不存在")
)
