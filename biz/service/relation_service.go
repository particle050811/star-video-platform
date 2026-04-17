package service

import (
	"context"
	"errors"
	"strconv"
	relation "video-platform/biz/model/relation"
	"video-platform/biz/repository"
	"video-platform/pkg/pagination"

	"gorm.io/gorm"
)

const (
	relationActionFollow   = relation.RelationActionType_RELATION_ACTION_TYPE_FOLLOW
	relationActionUnfollow = relation.RelationActionType_RELATION_ACTION_TYPE_UNFOLLOW
)

func RelationAction(ctx context.Context, fromUserID, toUserID uint, actionType relation.RelationActionType) error {
	if fromUserID == toUserID {
		return ErrCannotFollowSelf
	}

	if _, err := repository.GetUserByID(ctx, toUserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	switch actionType {
	case relationActionFollow:
		if err := repository.FollowUser(ctx, fromUserID, toUserID); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrAlreadyFollowed
			}
			return err
		}
		return nil
	case relationActionUnfollow:
		deleted, err := repository.UnfollowUser(ctx, fromUserID, toUserID)
		if err != nil {
			return err
		}
		if !deleted {
			return ErrFollowNotFound
		}
		return nil
	}

	return nil
}

func ListFollowings(ctx context.Context, userID uint, pageNum, pageSize int32) (*relation.SocialListWithTotal, error) {
	if _, err := repository.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	offset, limit := pagination.Normalize(pageNum, pageSize)
	users, total, err := repository.ListFollowings(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	return buildSocialList(users, total), nil
}

func ListFollowers(ctx context.Context, userID uint, pageNum, pageSize int32) (*relation.SocialListWithTotal, error) {
	if _, err := repository.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	offset, limit := pagination.Normalize(pageNum, pageSize)
	users, total, err := repository.ListFollowers(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	return buildSocialList(users, total), nil
}

func ListFriends(ctx context.Context, userID uint, pageNum, pageSize int32) (*relation.SocialListWithTotal, error) {
	if _, err := repository.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	offset, limit := pagination.Normalize(pageNum, pageSize)
	users, total, err := repository.ListFriends(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	return buildSocialList(users, total), nil
}

func buildSocialList(users []repository.UserSnapshot, total int64) *relation.SocialListWithTotal {
	items := make([]*relation.SocialProfile, 0, len(users))
	for _, user := range users {
		items = append(items, &relation.SocialProfile{
			Id:        strconv.FormatUint(uint64(user.ID), 10),
			Username:  user.Username,
			AvatarUrl: user.AvatarURL,
		})
	}

	return &relation.SocialListWithTotal{
		Items: items,
		Total: total,
	}
}
