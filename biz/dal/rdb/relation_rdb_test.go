package rdb

import (
	"context"
	"encoding/json"
	"testing"

	redismock "github.com/go-redis/redismock/v9"
)

func TestRelationCacheGetRelationFollowingCache(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewRelationCache(client)

	payload := RelationIDListCache{
		UserIDs: []uint{2, 3},
		Total:   2,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectGet("relation:following:1:v4:0:20").SetVal(string(raw))

	got, ok, err := cache.GetRelationFollowingCache(context.Background(), 1, 4, 0, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Total != 2 || len(got.UserIDs) != 2 || got.UserIDs[0] != 2 {
		t.Fatalf("unexpected payload: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestRelationCacheSetRelationFriendCache(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewRelationCache(client)

	payload := RelationIDListCache{
		UserIDs: []uint{4},
		Total:   1,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	mock.ExpectSet("relation:friend:1:v2:0:20", raw, relationCacheTTL).SetVal("OK")

	if err := cache.SetRelationFriendCache(context.Background(), 1, 2, 0, 20, payload); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}

func TestDefaultRelationCacheUsesGlobalRedisClient(t *testing.T) {
	client, mock := redismock.NewClientMock()
	prev := RDB
	RDB = client
	defer func() {
		RDB = prev
	}()

	mock.ExpectIncr("relation:following:version:1").SetVal(2)

	if err := Relations.BumpFollowingCacheVersion(context.Background(), 1); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet redis expectations: %v", err)
	}
}
