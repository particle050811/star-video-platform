package rdb

import (
	"context"
	"encoding/json"
	"testing"
	"video-platform/biz/dal/model"

	redismock "github.com/go-redis/redismock/v9"
)

func TestVideoCacheGetVideoDetailCache(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewVideoCache(client)

	payload := model.Video{
		ID:    1,
		Title: "cached video",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectGet("video:detail:1").SetVal(string(raw))

	var got model.Video
	ok, err := cache.GetVideoDetailCache(context.Background(), 1, &got)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Title != "cached video" {
		t.Fatalf("expected title %q, got %q", "cached video", got.Title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestVideoCacheSetHotVideoCache(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewVideoCache(client)

	payload := []model.Video{{ID: 1, Title: "hot"}}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectSet("video:hot:v3:0:20", raw, hotVideoCacheTTL).SetVal("OK")

	if err := cache.SetHotVideoCache(context.Background(), 3, 0, 20, payload); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestDefaultVideoCacheUsesGlobalRedisClient(t *testing.T) {
	client, mock := redismock.NewClientMock()
	prev := RDB
	RDB = client
	defer func() {
		RDB = prev
	}()

	payload := []model.Video{{ID: 2, Title: "global hot"}}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectGet("video:hot:v1:0:20").SetVal(string(raw))

	var got []model.Video
	ok, err := DefaultVideoCache.GetHotVideoCache(context.Background(), 1, 0, 20, &got)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 1 || got[0].Title != "global hot" {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}
