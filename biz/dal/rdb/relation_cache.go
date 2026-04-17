package rdb

import (
	"context"
	"fmt"
	"time"
)

const relationCacheTTL = 3 * time.Minute

type RelationIDListCache struct {
	UserIDs []uint `json:"user_ids"`
	Total   int64  `json:"total"`
}

func relationFollowingCacheKey(userID uint, offset, limit int) string {
	return fmt.Sprintf("relation:following:%d:%d:%d", userID, offset, limit)
}

func relationFollowerCacheKey(userID uint, offset, limit int) string {
	return fmt.Sprintf("relation:follower:%d:%d:%d", userID, offset, limit)
}

func relationFriendCacheKey(userID uint, offset, limit int) string {
	return fmt.Sprintf("relation:friend:%d:%d:%d", userID, offset, limit)
}

func GetRelationFollowingCache(ctx context.Context, userID uint, offset, limit int) (*RelationIDListCache, bool, error) {
	var cache RelationIDListCache
	ok, err := getJSON(ctx, relationFollowingCacheKey(userID, offset, limit), &cache)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &cache, true, nil
}

func GetRelationFollowerCache(ctx context.Context, userID uint, offset, limit int) (*RelationIDListCache, bool, error) {
	var cache RelationIDListCache
	ok, err := getJSON(ctx, relationFollowerCacheKey(userID, offset, limit), &cache)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &cache, true, nil
}

func GetRelationFriendCache(ctx context.Context, userID uint, offset, limit int) (*RelationIDListCache, bool, error) {
	var cache RelationIDListCache
	ok, err := getJSON(ctx, relationFriendCacheKey(userID, offset, limit), &cache)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &cache, true, nil
}

func SetRelationFollowingCache(ctx context.Context, userID uint, offset, limit int, value any) error {
	return setJSON(ctx, relationFollowingCacheKey(userID, offset, limit), value, relationCacheTTL)
}

func SetRelationFollowerCache(ctx context.Context, userID uint, offset, limit int, value any) error {
	return setJSON(ctx, relationFollowerCacheKey(userID, offset, limit), value, relationCacheTTL)
}

func SetRelationFriendCache(ctx context.Context, userID uint, offset, limit int, value any) error {
	return setJSON(ctx, relationFriendCacheKey(userID, offset, limit), value, relationCacheTTL)
}

func DeleteFollowingCaches(ctx context.Context, userID uint) error {
	return deleteByPattern(ctx, fmt.Sprintf("relation:following:%d:*", userID))
}

func DeleteFollowerCaches(ctx context.Context, userID uint) error {
	return deleteByPattern(ctx, fmt.Sprintf("relation:follower:%d:*", userID))
}

func DeleteFriendCaches(ctx context.Context, userID uint) error {
	return deleteByPattern(ctx, fmt.Sprintf("relation:friend:%d:*", userID))
}
