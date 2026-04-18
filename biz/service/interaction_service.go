package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"video-platform/biz/dal/model"
	interaction "video-platform/biz/model/interaction"
	videomodel "video-platform/biz/model/video"
	"video-platform/biz/repository"
	"video-platform/pkg/pagination"

	"gorm.io/gorm"
)

const (
	likeActionAdd    = interaction.LikeActionType_LIKE_ACTION_TYPE_ADD
	likeActionCancel = interaction.LikeActionType_LIKE_ACTION_TYPE_CANCEL
	maxCommentLength = 500
)

type interactionRepository interface {
	GetUserByID(ctx context.Context, userID uint) (*repository.UserProfile, error)
	GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error)
	LikeVideo(ctx context.Context, userID, videoID uint) error
	CancelLikeVideo(ctx context.Context, userID, videoID uint) (bool, error)
	ListLikedVideos(ctx context.Context, userID uint, cursor uint, limit int) (*repository.LikedVideoListResult, error)
	CreateComment(ctx context.Context, comment *model.Comment) error
	ListUserComments(ctx context.Context, userID uint, cursor uint, limit int) (*repository.UserCommentListResult, error)
	GetCommentByID(ctx context.Context, commentID uint) (*model.Comment, error)
	DeleteComment(ctx context.Context, commentID uint) error
}

type defaultInteractionRepository struct{}

func (defaultInteractionRepository) GetUserByID(ctx context.Context, userID uint) (*repository.UserProfile, error) {
	return repository.GetUserByID(ctx, userID)
}

func (defaultInteractionRepository) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return repository.GetVideoByID(ctx, videoID)
}

func (defaultInteractionRepository) LikeVideo(ctx context.Context, userID, videoID uint) error {
	return repository.LikeVideo(ctx, userID, videoID)
}

func (defaultInteractionRepository) CancelLikeVideo(ctx context.Context, userID, videoID uint) (bool, error) {
	return repository.CancelLikeVideo(ctx, userID, videoID)
}

func (defaultInteractionRepository) ListLikedVideos(ctx context.Context, userID uint, cursor uint, limit int) (*repository.LikedVideoListResult, error) {
	return repository.ListLikedVideos(ctx, userID, cursor, limit)
}

func (defaultInteractionRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	return repository.CreateComment(ctx, comment)
}

func (defaultInteractionRepository) ListUserComments(ctx context.Context, userID uint, cursor uint, limit int) (*repository.UserCommentListResult, error) {
	return repository.ListUserComments(ctx, userID, cursor, limit)
}

func (defaultInteractionRepository) GetCommentByID(ctx context.Context, commentID uint) (*model.Comment, error) {
	return repository.GetCommentByID(ctx, commentID)
}

func (defaultInteractionRepository) DeleteComment(ctx context.Context, commentID uint) error {
	return repository.DeleteComment(ctx, commentID)
}

type interactionService struct {
	repo interactionRepository
}

var Interaction = interactionService{
	repo: defaultInteractionRepository{},
}

func (s interactionService) VideoLikeAction(ctx context.Context, userID, videoID uint, actionType interaction.LikeActionType) error {
	if _, err := s.repo.GetVideoByID(ctx, videoID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVideoNotFound
		}
		return err
	}

	switch actionType {
	case likeActionAdd:
		if err := s.repo.LikeVideo(ctx, userID, videoID); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrAlreadyLiked
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrVideoNotFound
			}
			return err
		}
	case likeActionCancel:
		deleted, err := s.repo.CancelLikeVideo(ctx, userID, videoID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrVideoNotFound
			}
			return err
		}
		if !deleted {
			return ErrLikeNotFound
		}
	}

	return nil
}

func (s interactionService) ListLikedVideos(ctx context.Context, userID uint, cursor uint, limit int32) (*videomodel.VideoList, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	result, err := s.repo.ListLikedVideos(ctx, userID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}

	return buildVideoList(&repository.VideoListResult{
		Items:      result.Items,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	}), nil
}

func (s interactionService) PublishComment(ctx context.Context, userID, videoID uint, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return ErrCommentEmpty
	}
	if len([]rune(content)) > maxCommentLength {
		return ErrCommentTooLong
	}

	if _, err := s.repo.GetVideoByID(ctx, videoID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVideoNotFound
		}
		return err
	}

	if err := s.repo.CreateComment(ctx, &model.Comment{
		UserID:  userID,
		VideoID: videoID,
		Content: content,
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVideoNotFound
		}
		return err
	}

	return nil
}

func (s interactionService) ListUserComments(ctx context.Context, userID uint, cursor uint, limit int32) (*interaction.UserCommentList, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	result, err := s.repo.ListUserComments(ctx, userID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}

	return buildUserCommentList(result), nil
}

func buildUserCommentList(result *repository.UserCommentListResult) *interaction.UserCommentList {
	items := make([]*interaction.UserComment, 0)
	if result == nil {
		return &interaction.UserCommentList{
			Items: items,
		}
	}

	items = make([]*interaction.UserComment, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, &interaction.UserComment{
			Id:        strconv.FormatUint(uint64(item.ID), 10),
			UserId:    strconv.FormatUint(uint64(item.UserID), 10),
			VideoId:   strconv.FormatUint(uint64(item.VideoID), 10),
			Content:   item.Content,
			LikeCount: item.LikeCount,
			CreatedAt: item.CreatedAt.Format(time.RFC3339),
		})
	}

	nextCursor := ""
	if result.HasMore {
		nextCursor = strconv.FormatUint(uint64(result.NextCursor), 10)
	}

	return &interaction.UserCommentList{
		Items:      items,
		Total:      result.Total,
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	}
}

func (s interactionService) DeleteComment(ctx context.Context, userID, commentID uint) error {
	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		return err
	}

	if comment.UserID != userID {
		return ErrNoPermission
	}

	if err := s.repo.DeleteComment(ctx, commentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		return err
	}

	return nil
}
