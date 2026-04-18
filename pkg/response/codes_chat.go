package response

// 聊天模块响应码
const (
	CodeChatRoomNotFound       = 1021
	CodeChatRoomMemberNotFound = 1022
	CodeChatMemberExists       = 1023
	CodeChatMemberRequired     = 1024
	CodeChatPrivateMemberCount = 1025
	CodeChatGroupNameRequired  = 1026
	CodeChatOwnerCannotLeave   = 1027
	CodeChatMessageNotFound    = 1028
)

var chatCodeMsg = map[int32]string{
	CodeChatRoomNotFound:       "聊天房间不存在",
	CodeChatRoomMemberNotFound: "不是聊天房间成员",
	CodeChatMemberExists:       "用户已在聊天房间中",
	CodeChatMemberRequired:     "聊天房间成员不能为空",
	CodeChatPrivateMemberCount: "私聊房间只能包含两个成员",
	CodeChatGroupNameRequired:  "群聊名称不能为空",
	CodeChatOwnerCannotLeave:   "群主不能直接退出群聊",
	CodeChatMessageNotFound:    "聊天消息不存在",
}
