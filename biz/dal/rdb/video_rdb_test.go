package rdb

import (
	"context"
	"encoding/json"
	"testing"
	"video-platform/biz/dal/model"
	"video-platform/pkg/parser"

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

	mock.ExpectSet("video:hot:v3:0:0:0:20", raw, hotVideoCacheTTL).SetVal("OK")

	if err := cache.SetHotVideoCache(context.Background(), 3, parser.HotVideoCursorValue{}, 20, payload); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestVideoCacheSetHotVideoCacheWithCompositeCursor(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewVideoCache(client)

	payload := []model.Video{{ID: 3, Title: "hot page"}}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	cursor := parser.HotVideoCursorValue{LikeCount: 100, VisitCount: 200, ID: 9}
	mock.ExpectSet("video:hot:v4:100:200:9:10", raw, hotVideoCacheTTL).SetVal("OK")

	if err := cache.SetHotVideoCache(context.Background(), 4, cursor, 10, payload); err != nil {
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

	mock.ExpectGet("video:hot:v1:0:0:0:20").SetVal(string(raw))

	var got []model.Video
	ok, err := DefaultVideoCache.GetHotVideoCache(context.Background(), 1, parser.HotVideoCursorValue{}, 20, &got)
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
