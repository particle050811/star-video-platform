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

func relationFollowingCacheVersionKey(userID uint) string {
	return fmt.Sprintf("relation:following:version:%d", userID)
}

func relationFollowingCacheKey(userID uint, version int64, offset, limit int) string {
	return fmt.Sprintf("relation:following:%d:v%d:%d:%d", userID, version, offset, limit)
}

func relationFollowerCacheVersionKey(userID uint) string {
	return fmt.Sprintf("relation:follower:version:%d", userID)
}

func relationFollowerCacheKey(userID uint, version int64, offset, limit int) string {
	return fmt.Sprintf("relation:follower:%d:v%d:%d:%d", userID, version, offset, limit)
}

func relationFriendCacheVersionKey(userID uint) string {
	return fmt.Sprintf("relation:friend:version:%d", userID)
}

func relationFriendCacheKey(userID uint, version int64, offset, limit int) string {
	return fmt.Sprintf("relation:friend:%d:v%d:%d:%d", userID, version, offset, limit)
}

func GetRelationFollowingCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return getCacheVersion(ctx, relationFollowingCacheVersionKey(userID))
}

func GetRelationFollowingCache(ctx context.Context, userID uint, version int64, offset, limit int) (*RelationIDListCache, bool, error) {
	var cache RelationIDListCache
	ok, err := getJSON(ctx, relationFollowingCacheKey(userID, version, offset, limit), &cache)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &cache, true, nil
}

func GetRelationFollowerCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return getCacheVersion(ctx, relationFollowerCacheVersionKey(userID))
}

func GetRelationFollowerCache(ctx context.Context, userID uint, version int64, offset, limit int) (*RelationIDListCache, bool, error) {
	var cache RelationIDListCache
	ok, err := getJSON(ctx, relationFollowerCacheKey(userID, version, offset, limit), &cache)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &cache, true, nil
}

func GetRelationFriendCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return getCacheVersion(ctx, relationFriendCacheVersionKey(userID))
}

func GetRelationFriendCache(ctx context.Context, userID uint, version int64, offset, limit int) (*RelationIDListCache, bool, error) {
	var cache RelationIDListCache
	ok, err := getJSON(ctx, relationFriendCacheKey(userID, version, offset, limit), &cache)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &cache, true, nil
}

func SetRelationFollowingCache(ctx context.Context, userID uint, version int64, offset, limit int, value any) error {
	return setJSON(ctx, relationFollowingCacheKey(userID, version, offset, limit), value, relationCacheTTL)
}

func SetRelationFollowerCache(ctx context.Context, userID uint, version int64, offset, limit int, value any) error {
	return setJSON(ctx, relationFollowerCacheKey(userID, version, offset, limit), value, relationCacheTTL)
}

func SetRelationFriendCache(ctx context.Context, userID uint, version int64, offset, limit int, value any) error {
	return setJSON(ctx, relationFriendCacheKey(userID, version, offset, limit), value, relationCacheTTL)
}

func BumpFollowingCacheVersion(ctx context.Context, userID uint) error {
	return bumpCacheVersion(ctx, relationFollowingCacheVersionKey(userID))
}

func BumpFollowerCacheVersion(ctx context.Context, userID uint) error {
	return bumpCacheVersion(ctx, relationFollowerCacheVersionKey(userID))
}

func BumpFriendCacheVersion(ctx context.Context, userID uint) error {
	return bumpCacheVersion(ctx, relationFriendCacheVersionKey(userID))
}
