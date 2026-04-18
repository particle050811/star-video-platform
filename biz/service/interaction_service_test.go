package service

import (
	"context"
	"errors"
	"testing"
	"video-platform/biz/dal/model"
	interaction "video-platform/biz/model/interaction"
	"video-platform/biz/repository"

	"gorm.io/gorm"
)

type fakeInteractionRepository struct {
	getUserByIDFn      func(ctx context.Context, userID uint) (*repository.UserProfile, error)
	getVideoByIDFn     func(ctx context.Context, videoID uint) (*model.Video, error)
	likeVideoFn        func(ctx context.Context, userID, videoID uint) error
	cancelLikeVideoFn  func(ctx context.Context, userID, videoID uint) (bool, error)
	listLikedVideosFn  func(ctx context.Context, userID uint, cursor uint, limit int) (*repository.LikedVideoListResult, error)
	createCommentFn    func(ctx context.Context, comment *model.Comment) error
	listUserCommentsFn func(ctx context.Context, userID uint, cursor uint, limit int) (*repository.UserCommentListResult, error)
	getCommentByIDFn   func(ctx context.Context, commentID uint) (*model.Comment, error)
	deleteCommentFn    func(ctx context.Context, commentID uint) error
}

func (f fakeInteractionRepository) GetUserByID(ctx context.Context, userID uint) (*repository.UserProfile, error) {
	return f.getUserByIDFn(ctx, userID)
}

func (f fakeInteractionRepository) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return f.getVideoByIDFn(ctx, videoID)
}

func (f fakeInteractionRepository) LikeVideo(ctx context.Context, userID, videoID uint) error {
	return f.likeVideoFn(ctx, userID, videoID)
}

func (f fakeInteractionRepository) CancelLikeVideo(ctx context.Context, userID, videoID uint) (bool, error) {
	return f.cancelLikeVideoFn(ctx, userID, videoID)
}

func (f fakeInteractionRepository) ListLikedVideos(ctx context.Context, userID uint, cursor uint, limit int) (*repository.LikedVideoListResult, error) {
	return f.listLikedVideosFn(ctx, userID, cursor, limit)
}

func (f fakeInteractionRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	return f.createCommentFn(ctx, comment)
}

func (f fakeInteractionRepository) ListUserComments(ctx context.Context, userID uint, cursor uint, limit int) (*repository.UserCommentListResult, error) {
	return f.listUserCommentsFn(ctx, userID, cursor, limit)
}

func (f fakeInteractionRepository) GetCommentByID(ctx context.Context, commentID uint) (*model.Comment, error) {
	return f.getCommentByIDFn(ctx, commentID)
}

func (f fakeInteractionRepository) DeleteComment(ctx context.Context, commentID uint) error {
	return f.deleteCommentFn(ctx, commentID)
}

func TestInteractionServiceVideoLikeActionMapsDuplicateLike(t *testing.T) {
	svc := interactionService{
		repo: fakeInteractionRepository{
			getVideoByIDFn: func(ctx context.Context, videoID uint) (*model.Video, error) {
				return &model.Video{ID: videoID}, nil
			},
			likeVideoFn: func(ctx context.Context, userID, videoID uint) error {
				return gorm.ErrDuplicatedKey
			},
		},
	}

	err := svc.VideoLikeAction(context.Background(), 1, 2, interaction.LikeActionType_LIKE_ACTION_TYPE_ADD)
	if !errors.Is(err, ErrAlreadyLiked) {
		t.Fatalf("expected error %v, got %v", ErrAlreadyLiked, err)
	}
}

func TestInteractionServiceVideoLikeActionMapsConcurrentVideoDelete(t *testing.T) {
	svc := interactionService{
		repo: fakeInteractionRepository{
			getVideoByIDFn: func(ctx context.Context, videoID uint) (*model.Video, error) {
				return &model.Video{ID: videoID}, nil
			},
			likeVideoFn: func(ctx context.Context, userID, videoID uint) error {
				return gorm.ErrRecordNotFound
			},
		},
	}

	err := svc.VideoLikeAction(context.Background(), 1, 2, interaction.LikeActionType_LIKE_ACTION_TYPE_ADD)
	if !errors.Is(err, ErrVideoNotFound) {
		t.Fatalf("expected error %v, got %v", ErrVideoNotFound, err)
	}
}

func TestInteractionServicePublishCommentValidatesContent(t *testing.T) {
	svc := interactionService{}

	if err := svc.PublishComment(context.Background(), 1, 2, "   "); !errors.Is(err, ErrCommentEmpty) {
		t.Fatalf("expected error %v, got %v", ErrCommentEmpty, err)
	}
}

func TestInteractionServicePublishCommentMapsConcurrentVideoDelete(t *testing.T) {
	svc := interactionService{
		repo: fakeInteractionRepository{
			getVideoByIDFn: func(ctx context.Context, videoID uint) (*model.Video, error) {
				return &model.Video{ID: videoID}, nil
			},
			createCommentFn: func(ctx context.Context, comment *model.Comment) error {
				return gorm.ErrRecordNotFound
			},
		},
	}

	err := svc.PublishComment(context.Background(), 1, 2, "hello")
	if !errors.Is(err, ErrVideoNotFound) {
		t.Fatalf("expected error %v, got %v", ErrVideoNotFound, err)
	}
}

func TestInteractionServiceDeleteCommentChecksPermission(t *testing.T) {
	svc := interactionService{
		repo: fakeInteractionRepository{
			getCommentByIDFn: func(ctx context.Context, commentID uint) (*model.Comment, error) {
				return &model.Comment{ID: commentID, UserID: 2}, nil
			},
		},
	}

	err := svc.DeleteComment(context.Background(), 1, 3)
	if !errors.Is(err, ErrNoPermission) {
		t.Fatalf("expected error %v, got %v", ErrNoPermission, err)
	}
}

func TestInteractionServiceDeleteCommentMapsConcurrentDeleteToNotFound(t *testing.T) {
	svc := interactionService{
		repo: fakeInteractionRepository{
			getCommentByIDFn: func(ctx context.Context, commentID uint) (*model.Comment, error) {
				return &model.Comment{ID: commentID, UserID: 1}, nil
			},
			deleteCommentFn: func(ctx context.Context, commentID uint) error {
				return gorm.ErrRecordNotFound
			},
		},
	}

	err := svc.DeleteComment(context.Background(), 1, 3)
	if !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("expected error %v, got %v", ErrCommentNotFound, err)
	}
}

func TestInteractionServiceListUserCommentsBuildsCursorResponse(t *testing.T) {
	svc := interactionService{
		repo: fakeInteractionRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*repository.UserProfile, error) {
				return &repository.UserProfile{ID: userID}, nil
			},
			listUserCommentsFn: func(ctx context.Context, userID uint, cursor uint, limit int) (*repository.UserCommentListResult, error) {
				if userID != 5 || cursor != 10 || limit != 20 {
					t.Fatalf("unexpected params user=%d cursor=%d limit=%d", userID, cursor, limit)
				}
				return &repository.UserCommentListResult{
					Items: []repository.UserCommentItem{{
						ID:        9,
						UserID:    5,
						VideoID:   7,
						Content:   "hello",
						LikeCount: 3,
					}},
					Total:      1,
					NextCursor: 9,
					HasMore:    true,
				}, nil
			},
		},
	}

	got, err := svc.ListUserComments(context.Background(), 5, 10, 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Total != 1 || len(got.Items) != 1 {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.Items[0].Id != "9" || got.Items[0].VideoId != "7" {
		t.Fatalf("unexpected first item: %+v", got.Items[0])
	}
	if got.NextCursor != "9" || !got.HasMore {
		t.Fatalf("unexpected cursor result: %+v", got)
	}
}

func TestInteractionServiceListUserCommentsReturnsEmptyList(t *testing.T) {
	svc := interactionService{
		repo: fakeInteractionRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*repository.UserProfile, error) {
				return &repository.UserProfile{ID: userID}, nil
			},
			listUserCommentsFn: func(ctx context.Context, userID uint, cursor uint, limit int) (*repository.UserCommentListResult, error) {
				return &repository.UserCommentListResult{}, nil
			},
		},
	}

	got, err := svc.ListUserComments(context.Background(), 5, 0, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil data")
	}
	if got.Items == nil {
		t.Fatal("expected non-nil empty items")
	}
	if len(got.Items) != 0 || got.Total != 0 || got.NextCursor != "" || got.HasMore {
		t.Fatalf("unexpected empty result: %+v", got)
	}
}

func TestBuildUserCommentListHandlesNilResult(t *testing.T) {
	got := buildUserCommentList(nil)
	if got == nil {
		t.Fatal("expected non-nil data")
	}
	if got.Items == nil {
		t.Fatal("expected non-nil empty items")
	}
	if len(got.Items) != 0 {
		t.Fatalf("expected empty items, got %+v", got.Items)
	}
}
