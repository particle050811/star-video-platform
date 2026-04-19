package relation

import (
	"context"
	"errors"
	"strconv"
	relation "video-platform/biz/model/relation"
	relationrepo "video-platform/biz/repository/relation"
	userrepo "video-platform/biz/repository/user"
	usersvc "video-platform/biz/service/user"
	"video-platform/pkg/pagination"

	"gorm.io/gorm"
)

const (
	relationActionFollow   = relation.RelationActionType_RELATION_ACTION_TYPE_FOLLOW
	relationActionUnfollow = relation.RelationActionType_RELATION_ACTION_TYPE_UNFOLLOW
)

type relationRepository interface {
	GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error)
	FollowUser(ctx context.Context, fromUserID, toUserID uint) error
	UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error)
	ListFollowings(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error)
	ListFollowers(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error)
	ListFriends(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error)
}

type defaultRelationRepository struct{}

func (defaultRelationRepository) GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
	return userrepo.GetUserByID(ctx, userID)
}

func (defaultRelationRepository) FollowUser(ctx context.Context, fromUserID, toUserID uint) error {
	return relationrepo.FollowUser(ctx, fromUserID, toUserID)
}

func (defaultRelationRepository) UnfollowUser(ctx context.Context, fromUserID, toUserID uint) (bool, error) {
	return relationrepo.UnfollowUser(ctx, fromUserID, toUserID)
}

func (defaultRelationRepository) ListFollowings(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error) {
	return relationrepo.ListFollowings(ctx, userID, cursor, limit)
}

func (defaultRelationRepository) ListFollowers(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error) {
	return relationrepo.ListFollowers(ctx, userID, cursor, limit)
}

func (defaultRelationRepository) ListFriends(ctx context.Context, userID uint, cursor uint, limit int) (*relationrepo.RelationListResult, error) {
	return relationrepo.ListFriends(ctx, userID, cursor, limit)
}

type relationService struct {
	repo relationRepository
}

var defaultRelationService = relationService{
	repo: defaultRelationRepository{},
}

var Relation = defaultRelationService

func (s relationService) RelationAction(ctx context.Context, fromUserID, toUserID uint, actionType relation.RelationActionType) error {
	if actionType != relationActionFollow && actionType != relationActionUnfollow {
		return ErrInvalidRelationActionType
	}

	if fromUserID == toUserID {
		return ErrCannotFollowSelf
	}

	if _, err := s.repo.GetUserByID(ctx, toUserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return usersvc.ErrUserNotFound
		}
		return err
	}

	switch actionType {
	case relationActionFollow:
		if err := s.repo.FollowUser(ctx, fromUserID, toUserID); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrAlreadyFollowed
			}
			return err
		}
		return nil
	case relationActionUnfollow:
		deleted, err := s.repo.UnfollowUser(ctx, fromUserID, toUserID)
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

func (s relationService) ListFollowings(ctx context.Context, userID uint, cursor uint, limit int32) (*relation.SocialListWithTotal, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usersvc.ErrUserNotFound
		}
		return nil, err
	}

	result, err := s.repo.ListFollowings(ctx, userID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}

	return buildSocialList(result), nil
}

func (s relationService) ListFollowers(ctx context.Context, userID uint, cursor uint, limit int32) (*relation.SocialListWithTotal, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usersvc.ErrUserNotFound
		}
		return nil, err
	}

	result, err := s.repo.ListFollowers(ctx, userID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}

	return buildSocialList(result), nil
}

func (s relationService) ListFriends(ctx context.Context, userID uint, cursor uint, limit int32) (*relation.SocialListWithTotal, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usersvc.ErrUserNotFound
		}
		return nil, err
	}

	result, err := s.repo.ListFriends(ctx, userID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}

	return buildSocialList(result), nil
}

func buildSocialList(result *relationrepo.RelationListResult) *relation.SocialListWithTotal {
	items := make([]*relation.SocialProfile, 0)
	if result == nil {
		return &relation.SocialListWithTotal{
			Items: items,
		}
	}

	items = make([]*relation.SocialProfile, 0, len(result.Users))
	for _, user := range result.Users {
		items = append(items, &relation.SocialProfile{
			Id:        strconv.FormatUint(uint64(user.ID), 10),
			Username:  user.Username,
			AvatarUrl: user.AvatarURL,
		})
	}

	nextCursor := ""
	if result.HasMore {
		nextCursor = strconv.FormatUint(uint64(result.NextCursor), 10)
	}

	return &relation.SocialListWithTotal{
		Items:      items,
		Total:      result.Total,
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	}
}
