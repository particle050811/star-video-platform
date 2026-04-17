package service

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"
	"video-platform/biz/dal/model"
	v1 "video-platform/biz/model/video"
	"video-platform/biz/repository"
	"video-platform/pkg/upload"

	"gorm.io/gorm"
)

type fakeVideoRepository struct {
	createVideoFn           func(ctx context.Context, video *model.Video) error
	getUserByIDFn           func(ctx context.Context, userID uint) (*repository.UserProfile, error)
	listUserIDsByUsernameFn func(ctx context.Context, username string) ([]uint, error)
	getVideoByIDFn          func(ctx context.Context, videoID uint) (*model.Video, error)
	listVideosByUserIDFn    func(ctx context.Context, userID uint, cursor uint, limit int) (*repository.VideoListResult, error)
	searchVideosFn          func(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, cursor uint, limit int) (*repository.VideoListResult, error)
	listVideoCommentsFn     func(ctx context.Context, videoID uint, cursor uint, limit int) (*repository.VideoCommentListResult, error)
	listUserSnapshotsFn     func(ctx context.Context, userIDs []uint) ([]repository.UserProfile, error)
	listHotVideosFn         func(ctx context.Context, cursor uint, limit int) (*repository.VideoListResult, error)
}

func (f fakeVideoRepository) CreateVideo(ctx context.Context, video *model.Video) error {
	return f.createVideoFn(ctx, video)
}

func (f fakeVideoRepository) GetUserByID(ctx context.Context, userID uint) (*repository.UserProfile, error) {
	return f.getUserByIDFn(ctx, userID)
}

func (f fakeVideoRepository) ListUserIDsByUsername(ctx context.Context, username string) ([]uint, error) {
	return f.listUserIDsByUsernameFn(ctx, username)
}

func (f fakeVideoRepository) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return f.getVideoByIDFn(ctx, videoID)
}

func (f fakeVideoRepository) ListVideosByUserID(ctx context.Context, userID uint, cursor uint, limit int) (*repository.VideoListResult, error) {
	return f.listVideosByUserIDFn(ctx, userID, cursor, limit)
}

func (f fakeVideoRepository) SearchVideos(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, cursor uint, limit int) (*repository.VideoListResult, error) {
	return f.searchVideosFn(ctx, keywords, userIDs, fromDate, toDate, sortBy, cursor, limit)
}

func (f fakeVideoRepository) ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) (*repository.VideoCommentListResult, error) {
	return f.listVideoCommentsFn(ctx, videoID, cursor, limit)
}

func (f fakeVideoRepository) ListUserSnapshotsByIDs(ctx context.Context, userIDs []uint) ([]repository.UserProfile, error) {
	return f.listUserSnapshotsFn(ctx, userIDs)
}

func (f fakeVideoRepository) ListHotVideos(ctx context.Context, cursor uint, limit int) (*repository.VideoListResult, error) {
	return f.listHotVideosFn(ctx, cursor, limit)
}

type fakeUploadProvider struct {
	prepareAvatarFn     func(userID uint, originalFilename string) (savePath, avatarURL string, err error)
	saveFileFn          func(file *multipart.FileHeader, savePath string) error
	removeAvatarFn      func(avatarURL string) error
	prepareVideoFn      func(userID uint, originalFilename string) (savePath, videoURL string, err error)
	prepareVideoCoverFn func(userID uint, originalFilename string) (savePath, coverURL string, err error)
	removeVideoFn       func(videoURL string) error
	removeVideoCoverFn  func(coverURL string) error
}

func (f fakeUploadProvider) PrepareAvatar(userID uint, originalFilename string) (savePath, avatarURL string, err error) {
	if f.prepareAvatarFn == nil {
		return "", "", nil
	}
	return f.prepareAvatarFn(userID, originalFilename)
}

func (f fakeUploadProvider) SaveFile(file *multipart.FileHeader, savePath string) error {
	if f.saveFileFn == nil {
		return nil
	}
	return f.saveFileFn(file, savePath)
}

func (f fakeUploadProvider) RemoveAvatar(avatarURL string) error {
	if f.removeAvatarFn == nil {
		return nil
	}
	return f.removeAvatarFn(avatarURL)
}

func (f fakeUploadProvider) PrepareVideo(userID uint, originalFilename string) (savePath, videoURL string, err error) {
	if f.prepareVideoFn == nil {
		return "", "", nil
	}
	return f.prepareVideoFn(userID, originalFilename)
}

func (f fakeUploadProvider) PrepareVideoCover(userID uint, originalFilename string) (savePath, coverURL string, err error) {
	if f.prepareVideoCoverFn == nil {
		return "", "", nil
	}
	return f.prepareVideoCoverFn(userID, originalFilename)
}

func (f fakeUploadProvider) RemoveVideo(videoURL string) error {
	if f.removeVideoFn == nil {
		return nil
	}
	return f.removeVideoFn(videoURL)
}

func (f fakeUploadProvider) RemoveVideoCover(coverURL string) error {
	if f.removeVideoCoverFn == nil {
		return nil
	}
	return f.removeVideoCoverFn(coverURL)
}

func TestVideoServicePublishVideoValidation(t *testing.T) {
	svc := videoService{}

	err := svc.PublishVideo(context.Background(), 1, "", "desc", &multipart.FileHeader{Filename: "a.mp4"}, nil)
	if !errors.Is(err, ErrVideoTitleRequired) {
		t.Fatalf("expected error %v, got %v", ErrVideoTitleRequired, err)
	}

	err = svc.PublishVideo(context.Background(), 1, "title", "desc", nil, nil)
	if !errors.Is(err, ErrVideoFileRequired) {
		t.Fatalf("expected error %v, got %v", ErrVideoFileRequired, err)
	}
}

func TestVideoServicePublishVideoMapsUserNotFound(t *testing.T) {
	svc := videoService{
		repo: fakeVideoRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*repository.UserProfile, error) {
				return nil, gorm.ErrRecordNotFound
			},
		},
	}

	err := svc.PublishVideo(context.Background(), 1, "title", "desc", &multipart.FileHeader{Filename: "a.mp4"}, nil)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected error %v, got %v", ErrUserNotFound, err)
	}
}

func TestVideoServicePublishVideoMapsUnsupportedErrors(t *testing.T) {
	t.Run("unsupported video ext", func(t *testing.T) {
		svc := videoService{
			repo: fakeVideoRepository{
				getUserByIDFn: func(ctx context.Context, userID uint) (*repository.UserProfile, error) {
					return &repository.UserProfile{ID: userID}, nil
				},
			},
			upload: fakeUploadProvider{
				prepareVideoFn: func(userID uint, originalFilename string) (string, string, error) {
					return "", "", upload.ErrUnsupportedVideoExt
				},
			},
		}

		err := svc.PublishVideo(context.Background(), 1, "title", "desc", &multipart.FileHeader{Filename: "a.exe"}, nil)
		if !errors.Is(err, ErrUnsupportedVideoExt) {
			t.Fatalf("expected error %v, got %v", ErrUnsupportedVideoExt, err)
		}
	})

	t.Run("unsupported cover ext", func(t *testing.T) {
		svc := videoService{
			repo: fakeVideoRepository{
				getUserByIDFn: func(ctx context.Context, userID uint) (*repository.UserProfile, error) {
					return &repository.UserProfile{ID: userID}, nil
				},
			},
			upload: fakeUploadProvider{
				prepareVideoFn: func(userID uint, originalFilename string) (string, string, error) {
					return "/tmp/a.mp4", "/static/videos/a.mp4", nil
				},
				prepareVideoCoverFn: func(userID uint, originalFilename string) (string, string, error) {
					return "", "", upload.ErrUnsupportedVideoCoverExt
				},
			},
		}

		err := svc.PublishVideo(
			context.Background(),
			1,
			"title",
			"desc",
			&multipart.FileHeader{Filename: "a.mp4"},
			&multipart.FileHeader{Filename: "a.exe"},
		)
		if !errors.Is(err, ErrUnsupportedVideoCoverExt) {
			t.Fatalf("expected error %v, got %v", ErrUnsupportedVideoCoverExt, err)
		}
	})
}

func TestVideoServicePublishVideoRemovesPreparedFilesOnRepositoryError(t *testing.T) {
	var removedVideo string
	var removedCover string

	svc := videoService{
		repo: fakeVideoRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*repository.UserProfile, error) {
				return &repository.UserProfile{ID: userID}, nil
			},
			createVideoFn: func(ctx context.Context, video *model.Video) error {
				return errors.New("create video failed")
			},
		},
		upload: fakeUploadProvider{
			prepareVideoFn: func(userID uint, originalFilename string) (string, string, error) {
				return "/tmp/a.mp4", "/static/videos/a.mp4", nil
			},
			prepareVideoCoverFn: func(userID uint, originalFilename string) (string, string, error) {
				return "/tmp/a.jpg", "/static/video-covers/a.jpg", nil
			},
			saveFileFn: func(file *multipart.FileHeader, savePath string) error {
				return nil
			},
			removeVideoFn: func(videoURL string) error {
				removedVideo = videoURL
				return nil
			},
			removeVideoCoverFn: func(coverURL string) error {
				removedCover = coverURL
				return nil
			},
		},
	}

	err := svc.PublishVideo(
		context.Background(),
		1,
		"title",
		"desc",
		&multipart.FileHeader{Filename: "a.mp4"},
		&multipart.FileHeader{Filename: "a.jpg"},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if removedVideo != "/static/videos/a.mp4" {
		t.Fatalf("expected removed video %q, got %q", "/static/videos/a.mp4", removedVideo)
	}
	if removedCover != "/static/video-covers/a.jpg" {
		t.Fatalf("expected removed cover %q, got %q", "/static/video-covers/a.jpg", removedCover)
	}
}

func TestVideoServiceListPublishedVideosUsesCursor(t *testing.T) {
	svc := videoService{
		repo: fakeVideoRepository{
			getUserByIDFn: func(ctx context.Context, userID uint) (*repository.UserProfile, error) {
				return &repository.UserProfile{ID: userID}, nil
			},
			listVideosByUserIDFn: func(ctx context.Context, userID uint, cursor uint, limit int) (*repository.VideoListResult, error) {
				if userID != 7 {
					t.Fatalf("expected user id %d, got %d", 7, userID)
				}
				if cursor != 12 {
					t.Fatalf("expected cursor %d, got %d", 12, cursor)
				}
				if limit != 20 {
					t.Fatalf("expected limit %d, got %d", 20, limit)
				}
				return &repository.VideoListResult{
					Items:      []model.Video{{ID: 1, UserID: 7, Title: "video", CreatedAt: time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)}},
					NextCursor: 1,
					HasMore:    true,
				}, nil
			},
		},
	}

	got, err := svc.ListPublishedVideos(context.Background(), 7, 12, 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].Id != "1" {
		t.Fatalf("unexpected response: %+v", got)
	}
	if got.NextCursor != "1" || !got.HasMore {
		t.Fatalf("unexpected cursor response: %+v", got)
	}
}

func TestVideoServiceSearchVideosUsesUsernameFilter(t *testing.T) {
	svc := videoService{
		repo: fakeVideoRepository{
			listUserIDsByUsernameFn: func(ctx context.Context, username string) ([]uint, error) {
				if username != "alice" {
					t.Fatalf("expected username %q, got %q", "alice", username)
				}
				return []uint{3, 4}, nil
			},
			searchVideosFn: func(ctx context.Context, keywords string, userIDs []uint, fromDate, toDate int64, sortBy string, cursor uint, limit int) (*repository.VideoListResult, error) {
				if keywords != "golang" {
					t.Fatalf("expected keywords %q, got %q", "golang", keywords)
				}
				if len(userIDs) != 2 || userIDs[0] != 3 || userIDs[1] != 4 {
					t.Fatalf("unexpected userIDs: %+v", userIDs)
				}
				if cursor != 8 || limit != 20 {
					t.Fatalf("unexpected cursor pagination cursor=%d limit=%d", cursor, limit)
				}
				return &repository.VideoListResult{
					Items: []model.Video{{ID: 9, UserID: 3, Title: "result", CreatedAt: time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)}},
				}, nil
			},
		},
	}

	got, err := svc.SearchVideos(context.Background(), &v1.SearchVideosRequest{
		Keywords: "golang",
		Username: " alice ",
	}, 8)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].Id != "9" {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestVideoServiceListVideoCommentsBuildsCursorResponse(t *testing.T) {
	createdAt := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	svc := videoService{
		repo: fakeVideoRepository{
			getVideoByIDFn: func(ctx context.Context, videoID uint) (*model.Video, error) {
				return &model.Video{ID: videoID}, nil
			},
			listVideoCommentsFn: func(ctx context.Context, videoID uint, cursor uint, limit int) (*repository.VideoCommentListResult, error) {
				if cursor != 5 {
					t.Fatalf("expected cursor %d, got %d", 5, cursor)
				}
				if limit != 100 {
					t.Fatalf("expected limit %d, got %d", 100, limit)
				}
				return &repository.VideoCommentListResult{
					Items: []repository.VideoComment{
						{
							ID:        11,
							UserID:    22,
							Content:   "hello",
							LikeCount: 3,
							CreatedAt: createdAt,
						},
					},
					Total:      8,
					NextCursor: 11,
					HasMore:    true,
				}, nil
			},
			listUserSnapshotsFn: func(ctx context.Context, userIDs []uint) ([]repository.UserProfile, error) {
				if len(userIDs) != 1 || userIDs[0] != 22 {
					t.Fatalf("unexpected user IDs: %+v", userIDs)
				}
				return []repository.UserProfile{
					{ID: 22, Username: "alice", AvatarURL: "/static/avatars/a.png"},
				}, nil
			},
		},
	}

	got, err := svc.ListVideoComments(context.Background(), 1, 5, 1000)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Total != 8 {
		t.Fatalf("expected total %d, got %d", 8, got.Total)
	}
	if got.NextCursor != "11" {
		t.Fatalf("expected next cursor %q, got %q", "11", got.NextCursor)
	}
	if !got.HasMore {
		t.Fatal("expected has more to be true")
	}
	if len(got.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got.Items))
	}
	if got.Items[0].Username != "alice" {
		t.Fatalf("expected username %q, got %q", "alice", got.Items[0].Username)
	}
	if got.Items[0].CreatedAt != createdAt.Format(time.RFC3339) {
		t.Fatalf("expected created at %q, got %q", createdAt.Format(time.RFC3339), got.Items[0].CreatedAt)
	}
}
