package interaction

import "errors"

var (
	ErrCommentNotFound       = errors.New("评论不存在")
	ErrCommentTooLong        = errors.New("评论内容过长")
	ErrCommentEmpty          = errors.New("评论内容不能为空")
	ErrAlreadyLiked          = errors.New("已经点赞该视频")
	ErrLikeNotFound          = errors.New("点赞记录不存在")
	ErrInvalidLikeActionType = errors.New("非法点赞动作类型")
	ErrNoPermission          = errors.New("无权限操作")
)
