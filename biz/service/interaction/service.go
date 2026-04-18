package interaction

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"video-platform/biz/dal/model"
	interaction "video-platform/biz/model/interaction"
	videomodel "video-platform/biz/model/video"
	interactionrepo "video-platform/biz/repository/interaction"
	userrepo "video-platform/biz/repository/user"
	videorepo "video-platform/biz/repository/video"
	usersvc "video-platform/biz/service/user"
	videosvc "video-platform/biz/service/video"
	"video-platform/pkg/pagination"

	"gorm.io/gorm"
)

const (
	likeActionAdd    = interaction.LikeActionType_LIKE_ACTION_TYPE_ADD
	likeActionCancel = interaction.LikeActionType_LIKE_ACTION_TYPE_CANCEL
	maxCommentLength = 500
)

type interactionRepository interface {
	GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error)
	GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error)
	LikeVideo(ctx context.Context, userID, videoID uint) error
	CancelLikeVideo(ctx context.Context, userID, videoID uint) (bool, error)
	ListLikedVideos(ctx context.Context, userID uint, cursor uint, limit int) (*interactionrepo.LikedVideoListResult, error)
	CreateComment(ctx context.Context, comment *model.Comment) error
	ListUserComments(ctx context.Context, userID uint, cursor uint, limit int) (*interactionrepo.UserCommentListResult, error)
	GetCommentByID(ctx context.Context, commentID uint) (*model.Comment, error)
	DeleteComment(ctx context.Context, commentID uint) error
}

type defaultInteractionRepository struct{}

func (defaultInteractionRepository) GetUserByID(ctx context.Context, userID uint) (*userrepo.UserProfile, error) {
	return userrepo.GetUserByID(ctx, userID)
}

func (defaultInteractionRepository) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return videorepo.GetVideoByID(ctx, videoID)
}

func (defaultInteractionRepository) LikeVideo(ctx context.Context, userID, videoID uint) error {
	return interactionrepo.LikeVideo(ctx, userID, videoID)
}

func (defaultInteractionRepository) CancelLikeVideo(ctx context.Context, userID, videoID uint) (bool, error) {
	return interactionrepo.CancelLikeVideo(ctx, userID, videoID)
}

func (defaultInteractionRepository) ListLikedVideos(ctx context.Context, userID uint, cursor uint, limit int) (*interactionrepo.LikedVideoListResult, error) {
	return interactionrepo.ListLikedVideos(ctx, userID, cursor, limit)
}

func (defaultInteractionRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	return interactionrepo.CreateComment(ctx, comment)
}

func (defaultInteractionRepository) ListUserComments(ctx context.Context, userID uint, cursor uint, limit int) (*interactionrepo.UserCommentListResult, error) {
	return interactionrepo.ListUserComments(ctx, userID, cursor, limit)
}

func (defaultInteractionRepository) GetCommentByID(ctx context.Context, commentID uint) (*model.Comment, error) {
	return interactionrepo.GetCommentByID(ctx, commentID)
}

func (defaultInteractionRepository) DeleteComment(ctx context.Context, commentID uint) error {
	return interactionrepo.DeleteComment(ctx, commentID)
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
			return videosvc.ErrVideoNotFound
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
				return videosvc.ErrVideoNotFound
			}
			return err
		}
	case likeActionCancel:
		deleted, err := s.repo.CancelLikeVideo(ctx, userID, videoID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return videosvc.ErrVideoNotFound
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
			return nil, usersvc.ErrUserNotFound
		}
		return nil, err
	}

	result, err := s.repo.ListLikedVideos(ctx, userID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}

	return videosvc.BuildVideoList(&videorepo.VideoListResult{
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
			return videosvc.ErrVideoNotFound
		}
		return err
	}

	if err := s.repo.CreateComment(ctx, &model.Comment{
		UserID:  userID,
		VideoID: videoID,
		Content: content,
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return videosvc.ErrVideoNotFound
		}
		return err
	}

	return nil
}

func (s interactionService) ListUserComments(ctx context.Context, userID uint, cursor uint, limit int32) (*interaction.UserCommentList, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usersvc.ErrUserNotFound
		}
		return nil, err
	}

	result, err := s.repo.ListUserComments(ctx, userID, cursor, pagination.NormalizeLimit(limit))
	if err != nil {
		return nil, err
	}

	return buildUserCommentList(result), nil
}

func buildUserCommentList(result *interactionrepo.UserCommentListResult) *interaction.UserCommentList {
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
