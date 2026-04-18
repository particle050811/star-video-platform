package rdb

import (
	"context"
	"encoding/json"
	"testing"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

func TestChatCacheGetChatMessageCache(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewChatCache(client)

	payload := map[string]any{
		"next_cursor": float64(21),
		"has_more":    true,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectGet("chat:messages:6:v2:0:20").SetVal(string(raw))

	var got map[string]any
	ok, err := cache.GetChatMessageCache(context.Background(), 6, 2, 0, 20, &got)
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

func TestChatCacheSetChatMessageCache(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewChatCache(client)

	payload := map[string]any{
		"next_cursor": uint(21),
		"has_more":    true,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectSet("chat:messages:6:v2:0:20", raw, chatMessageCacheTTL).SetVal("OK")

	if err := cache.SetChatMessageCache(context.Background(), 6, 2, 0, 20, payload); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestDefaultChatCacheUsesGlobalRedisClient(t *testing.T) {
	client, mock := redismock.NewClientMock()
	prev := RDB
	RDB = client
	defer func() {
		RDB = prev
	}()

	mock.ExpectGet("chat:messages:version:6").SetVal("3")

	got, err := Chats.GetChatMessageCacheVersion(context.Background(), 6)
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

func TestChatCacheMissingVersionDefaultsToZero(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewChatCache(client)

	mock.ExpectGet("chat:messages:version:6").RedisNil()

	got, err := cache.GetChatMessageCacheVersion(context.Background(), 6)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 0 {
		t.Fatalf("expected version 0, got %d", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestChatCacheFirstBumpMovesVersionToOne(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewChatCache(client)

	mock.ExpectIncr("chat:messages:version:6").SetVal(1)

	if err := cache.BumpChatMessageCacheVersion(context.Background(), 6); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestChatCacheMissingMessageCacheReturnsMiss(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewChatCache(client)

	mock.ExpectGet("chat:messages:6:v0:0:20").SetErr(redis.Nil)

	var got map[string]any
	ok, err := cache.GetChatMessageCache(context.Background(), 6, 0, 0, 20, &got)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatal("expected cache miss")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}
