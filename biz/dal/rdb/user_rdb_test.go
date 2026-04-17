package rdb

import (
	"context"
	"encoding/json"
	"testing"

	redismock "github.com/go-redis/redismock/v9"
)

func TestUserCacheGetUserProfileCache(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewUserCache(client)

	payload := UserProfileCache{
		ID:             1,
		Username:       "alice",
		AvatarURL:      "/static/avatars/a.png",
		FollowingCount: 3,
		FollowerCount:  4,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectGet("user:profile:1").SetVal(string(raw))

	got, ok, err := cache.GetUserProfileCache(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Username != "alice" {
		t.Fatalf("expected username %q, got %q", "alice", got.Username)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestUserCacheSetUserProfileCache(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewUserCache(client)

	payload := UserProfileCache{
		ID:       2,
		Username: "bob",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectSet("user:profile:2", raw, userProfileCacheTTL).SetVal("OK")

	if err := cache.SetUserProfileCache(context.Background(), 2, payload); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestUserCacheDeleteUserProfileCache(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewUserCache(client)

	mock.ExpectDel("user:profile:3").SetVal(1)

	if err := cache.DeleteUserProfileCache(context.Background(), 3); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestDefaultUserCacheUsesGlobalRedisClient(t *testing.T) {
	client, mock := redismock.NewClientMock()
	prev := RDB
	RDB = client
	defer func() {
		RDB = prev
	}()

	payload := UserProfileCache{
		ID:       4,
		Username: "dora",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectGet("user:profile:4").SetVal(string(raw))

	got, ok, err := DefaultUserCache.GetUserProfileCache(context.Background(), 4)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Username != "dora" {
		t.Fatalf("expected username %q, got %q", "dora", got.Username)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}
