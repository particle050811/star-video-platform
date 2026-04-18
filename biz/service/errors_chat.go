package service

import "errors"

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
