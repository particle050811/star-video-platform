package service

import (
	chatsvc "video-platform/biz/service/chat"
	interactionsvc "video-platform/biz/service/interaction"
	relationsvc "video-platform/biz/service/relation"
	usersvc "video-platform/biz/service/user"
	videosvc "video-platform/biz/service/video"
)

var (
	User        = usersvc.User
	Video       = videosvc.Video
	Interaction = interactionsvc.Interaction
	Relation    = relationsvc.Relation
	Chat        = chatsvc.Chat
)

var (
	ErrUserExists           = usersvc.ErrUserExists
	ErrUserNotFound         = usersvc.ErrUserNotFound
	ErrPasswordWrong        = usersvc.ErrPasswordWrong
	ErrTokenExpired         = usersvc.ErrTokenExpired
	ErrTokenInvalid         = usersvc.ErrTokenInvalid
	ErrUnsupportedAvatarExt = usersvc.ErrUnsupportedAvatarExt

	ErrVideoNotFound            = videosvc.ErrVideoNotFound
	ErrVideoTitleRequired       = videosvc.ErrVideoTitleRequired
	ErrVideoFileRequired        = videosvc.ErrVideoFileRequired
	ErrUnsupportedVideoExt      = videosvc.ErrUnsupportedVideoExt
	ErrUnsupportedVideoCoverExt = videosvc.ErrUnsupportedVideoCoverExt
	ErrVideoCursorInvalid       = videosvc.ErrVideoCursorInvalid

	ErrCommentNotFound       = interactionsvc.ErrCommentNotFound
	ErrCommentTooLong        = interactionsvc.ErrCommentTooLong
	ErrCommentEmpty          = interactionsvc.ErrCommentEmpty
	ErrAlreadyLiked          = interactionsvc.ErrAlreadyLiked
	ErrLikeNotFound          = interactionsvc.ErrLikeNotFound
	ErrInvalidLikeActionType = interactionsvc.ErrInvalidLikeActionType
	ErrNoPermission          = interactionsvc.ErrNoPermission

	ErrCannotFollowSelf          = relationsvc.ErrCannotFollowSelf
	ErrAlreadyFollowed           = relationsvc.ErrAlreadyFollowed
	ErrFollowNotFound            = relationsvc.ErrFollowNotFound
	ErrInvalidRelationActionType = relationsvc.ErrInvalidRelationActionType

	ErrChatRoomNotFound       = chatsvc.ErrChatRoomNotFound
	ErrChatRoomMemberNotFound = chatsvc.ErrChatRoomMemberNotFound
	ErrChatMemberExists       = chatsvc.ErrChatMemberExists
	ErrChatMemberRequired     = chatsvc.ErrChatMemberRequired
	ErrChatPrivateMemberCount = chatsvc.ErrChatPrivateMemberCount
	ErrChatGroupNameRequired  = chatsvc.ErrChatGroupNameRequired
	ErrChatOwnerCannotLeave   = chatsvc.ErrChatOwnerCannotLeave
	ErrChatMessageNotFound    = chatsvc.ErrChatMessageNotFound
)
