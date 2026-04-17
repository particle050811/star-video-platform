package service

import (
	"testing"
	"time"
	"video-platform/biz/dal/model"
	"video-platform/biz/repository"

	"gorm.io/gorm"
)

func TestBuildVideo(t *testing.T) {
	createdAt := time.Date(2026, 4, 17, 10, 11, 12, 0, time.UTC)
	updatedAt := time.Date(2026, 4, 18, 13, 14, 15, 0, time.UTC)
	deletedAt := time.Date(2026, 4, 19, 16, 17, 18, 0, time.UTC)

	video := model.Video{
		ID:           1,
		UserID:       2,
		VideoURL:     "/static/videos/a.mp4",
		CoverURL:     "/static/video-covers/a.jpg",
		Title:        "video title",
		Description:  "video description",
		VisitCount:   30,
		LikeCount:    40,
		CommentCount: 50,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		DeletedAt: gorm.DeletedAt{
			Time:  deletedAt,
			Valid: true,
		},
	}

	got := buildVideo(video)
	if got.Id != "1" {
		t.Fatalf("expected id %q, got %q", "1", got.Id)
	}
	if got.UserId != "2" {
		t.Fatalf("expected user id %q, got %q", "2", got.UserId)
	}
	if got.VideoUrl != video.VideoURL {
		t.Fatalf("expected video url %q, got %q", video.VideoURL, got.VideoUrl)
	}
	if got.CoverUrl != video.CoverURL {
		t.Fatalf("expected cover url %q, got %q", video.CoverURL, got.CoverUrl)
	}
	if got.Title != video.Title {
		t.Fatalf("expected title %q, got %q", video.Title, got.Title)
	}
	if got.Description != video.Description {
		t.Fatalf("expected description %q, got %q", video.Description, got.Description)
	}
	if got.VisitCount != video.VisitCount {
		t.Fatalf("expected visit count %d, got %d", video.VisitCount, got.VisitCount)
	}
	if got.LikeCount != video.LikeCount {
		t.Fatalf("expected like count %d, got %d", video.LikeCount, got.LikeCount)
	}
	if got.CommentCount != video.CommentCount {
		t.Fatalf("expected comment count %d, got %d", video.CommentCount, got.CommentCount)
	}
	if got.CreatedAt != createdAt.Format(time.RFC3339) {
		t.Fatalf("expected created at %q, got %q", createdAt.Format(time.RFC3339), got.CreatedAt)
	}
	if got.UpdatedAt != updatedAt.Format(time.RFC3339) {
		t.Fatalf("expected updated at %q, got %q", updatedAt.Format(time.RFC3339), got.UpdatedAt)
	}
	if got.DeletedAt != deletedAt.Format(time.RFC3339) {
		t.Fatalf("expected deleted at %q, got %q", deletedAt.Format(time.RFC3339), got.DeletedAt)
	}
}

func TestBuildVideoWithoutDeletedAt(t *testing.T) {
	video := model.Video{
		ID:        7,
		UserID:    8,
		CreatedAt: time.Date(2026, 4, 17, 10, 11, 12, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 18, 13, 14, 15, 0, time.UTC),
	}

	got := buildVideo(video)
	if got.DeletedAt != "" {
		t.Fatalf("expected deleted at to be empty, got %q", got.DeletedAt)
	}
}

func TestBuildVideoList(t *testing.T) {
	videos := []model.Video{
		{
			ID:        1,
			UserID:    11,
			Title:     "first",
			CreatedAt: time.Date(2026, 4, 17, 10, 11, 12, 0, time.UTC),
			UpdatedAt: time.Date(2026, 4, 18, 13, 14, 15, 0, time.UTC),
		},
		{
			ID:        2,
			UserID:    22,
			Title:     "second",
			CreatedAt: time.Date(2026, 4, 19, 10, 11, 12, 0, time.UTC),
			UpdatedAt: time.Date(2026, 4, 20, 13, 14, 15, 0, time.UTC),
		},
	}

	got := buildVideoList(&repository.VideoListResult{
		Items:      videos,
		NextCursor: 2,
		HasMore:    true,
	})
	if len(got.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got.Items))
	}
	if got.Items[0].Id != "1" {
		t.Fatalf("expected first item id %q, got %q", "1", got.Items[0].Id)
	}
	if got.Items[1].Id != "2" {
		t.Fatalf("expected second item id %q, got %q", "2", got.Items[1].Id)
	}
	if got.Items[0].Title != "first" {
		t.Fatalf("expected first item title %q, got %q", "first", got.Items[0].Title)
	}
	if got.Items[1].Title != "second" {
		t.Fatalf("expected second item title %q, got %q", "second", got.Items[1].Title)
	}
	if got.NextCursor != "2" || !got.HasMore {
		t.Fatalf("unexpected cursor response: %+v", got)
	}
}
