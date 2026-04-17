package rdb

import (
	"context"
	"encoding/json"
	"testing"

	redismock "github.com/go-redis/redismock/v9"
)

func TestCommentCacheGetVideoCommentCache(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewCommentCache(client)

	payload := map[string]any{
		"total":       float64(2),
		"next_cursor": float64(3),
		"has_more":    true,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectGet("comment:list:9:v2:5:20").SetVal(string(raw))

	var got map[string]any
	ok, err := cache.GetVideoCommentCache(context.Background(), 9, 2, 5, 20, &got)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got["has_more"] != true {
		t.Fatalf("unexpected payload: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestCommentCacheSetVideoCommentCache(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewCommentCache(client)

	payload := map[string]any{
		"total": int64(1),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectSet("comment:list:9:v2:0:20", raw, videoCommentCacheTTL).SetVal("OK")

	if err := cache.SetVideoCommentCache(context.Background(), 9, 2, 0, 20, payload); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestDefaultCommentCacheUsesGlobalRedisClient(t *testing.T) {
	client, mock := redismock.NewClientMock()
	prev := RDB
	RDB = client
	defer func() {
		RDB = prev
	}()

	mock.ExpectGet("comment:list:version:9").SetVal("3")

	got, err := Comments.GetVideoCommentCacheVersion(context.Background(), 9)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 3 {
		t.Fatalf("expected version 3, got %d", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}
