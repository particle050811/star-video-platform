package rdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const relationCacheTTL = 3 * time.Minute

type RelationIDListCache struct {
	UserIDs    []uint `json:"user_ids"`
	Total      int64  `json:"total"`
	NextCursor uint   `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

type RelationCache struct {
	client *redis.Client
}

func NewRelationCache(client *redis.Client) RelationCache {
	return RelationCache{client: client}
}

var Relations = RelationCache{}

func relationFollowingCacheVersionKey(userID uint) string {
	return fmt.Sprintf("relation:following:version:%d", userID)
}

func relationFollowingCacheKey(userID uint, version int64, cursor uint, limit int) string {
	return fmt.Sprintf("relation:following:%d:v%d:%d:%d", userID, version, cursor, limit)
}

func relationFollowerCacheVersionKey(userID uint) string {
	return fmt.Sprintf("relation:follower:version:%d", userID)
}

func relationFollowerCacheKey(userID uint, version int64, cursor uint, limit int) string {
	return fmt.Sprintf("relation:follower:%d:v%d:%d:%d", userID, version, cursor, limit)
}

func relationFriendCacheVersionKey(userID uint) string {
	return fmt.Sprintf("relation:friend:version:%d", userID)
}

func relationFriendCacheKey(userID uint, version int64, cursor uint, limit int) string {
	return fmt.Sprintf("relation:friend:%d:v%d:%d:%d", userID, version, cursor, limit)
}

func (r RelationCache) redisClient() *redis.Client {
	if r.client != nil {
		return r.client
	}
	return RDB
}

func (r RelationCache) getCacheVersion(ctx context.Context, key string) (int64, error) {
	client := r.redisClient()
	if client == nil {
		return 1, nil
	}

	value, err := client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 1, nil
		}
		return 0, err
	}

	version, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	if version < 1 {
		return 0, errors.New("cache version must be greater than zero")
	}

	return version, nil
}

func (r RelationCache) getRelationIDListCache(ctx context.Context, key string) (*RelationIDListCache, bool, error) {
	client := r.redisClient()
	if client == nil {
		return nil, false, nil
	}

	var cache RelationIDListCache
	value, err := client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}

	if err := json.Unmarshal([]byte(value), &cache); err != nil {
		return nil, false, err
	}

	return &cache, true, nil
}

func (r RelationCache) setRelationIDListCache(ctx context.Context, key string, value any) error {
	client := r.redisClient()
	if client == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return client.Set(ctx, key, payload, relationCacheTTL).Err()
}

func (r RelationCache) bumpCacheVersion(ctx context.Context, key string) error {
	client := r.redisClient()
	if client == nil {
		return nil
	}

	return client.Incr(ctx, key).Err()
}

func (r RelationCache) GetRelationFollowingCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return r.getCacheVersion(ctx, relationFollowingCacheVersionKey(userID))
}

func (r RelationCache) GetRelationFollowingCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*RelationIDListCache, bool, error) {
	return r.getRelationIDListCache(ctx, relationFollowingCacheKey(userID, version, cursor, limit))
}

func (r RelationCache) GetRelationFollowerCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return r.getCacheVersion(ctx, relationFollowerCacheVersionKey(userID))
}

func (r RelationCache) GetRelationFollowerCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*RelationIDListCache, bool, error) {
	return r.getRelationIDListCache(ctx, relationFollowerCacheKey(userID, version, cursor, limit))
}

func (r RelationCache) GetRelationFriendCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return r.getCacheVersion(ctx, relationFriendCacheVersionKey(userID))
}

func (r RelationCache) GetRelationFriendCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*RelationIDListCache, bool, error) {
	return r.getRelationIDListCache(ctx, relationFriendCacheKey(userID, version, cursor, limit))
}

func (r RelationCache) SetRelationFollowingCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
	return r.setRelationIDListCache(ctx, relationFollowingCacheKey(userID, version, cursor, limit), value)
}

func (r RelationCache) SetRelationFollowerCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
	return r.setRelationIDListCache(ctx, relationFollowerCacheKey(userID, version, cursor, limit), value)
}

func (r RelationCache) SetRelationFriendCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
	return r.setRelationIDListCache(ctx, relationFriendCacheKey(userID, version, cursor, limit), value)
}

func (r RelationCache) BumpFollowingCacheVersion(ctx context.Context, userID uint) error {
	return r.bumpCacheVersion(ctx, relationFollowingCacheVersionKey(userID))
}

func (r RelationCache) BumpFollowerCacheVersion(ctx context.Context, userID uint) error {
	return r.bumpCacheVersion(ctx, relationFollowerCacheVersionKey(userID))
}

func (r RelationCache) BumpFriendCacheVersion(ctx context.Context, userID uint) error {
	return r.bumpCacheVersion(ctx, relationFriendCacheVersionKey(userID))
}

func (r RelationCache) DeleteUserProfileCache(ctx context.Context, userID uint) error {
	return DefaultUserCache.DeleteUserProfileCache(ctx, userID)
}

func GetRelationFollowingCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return Relations.GetRelationFollowingCacheVersion(ctx, userID)
}

func GetRelationFollowingCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*RelationIDListCache, bool, error) {
	return Relations.GetRelationFollowingCache(ctx, userID, version, cursor, limit)
}

func GetRelationFollowerCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return Relations.GetRelationFollowerCacheVersion(ctx, userID)
}

func GetRelationFollowerCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*RelationIDListCache, bool, error) {
	return Relations.GetRelationFollowerCache(ctx, userID, version, cursor, limit)
}

func GetRelationFriendCacheVersion(ctx context.Context, userID uint) (int64, error) {
	return Relations.GetRelationFriendCacheVersion(ctx, userID)
}

func GetRelationFriendCache(ctx context.Context, userID uint, version int64, cursor uint, limit int) (*RelationIDListCache, bool, error) {
	return Relations.GetRelationFriendCache(ctx, userID, version, cursor, limit)
}

func SetRelationFollowingCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
	return Relations.SetRelationFollowingCache(ctx, userID, version, cursor, limit, value)
}

func SetRelationFollowerCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
	return Relations.SetRelationFollowerCache(ctx, userID, version, cursor, limit, value)
}

func SetRelationFriendCache(ctx context.Context, userID uint, version int64, cursor uint, limit int, value any) error {
	return Relations.SetRelationFriendCache(ctx, userID, version, cursor, limit, value)
}

func BumpFollowingCacheVersion(ctx context.Context, userID uint) error {
	return Relations.BumpFollowingCacheVersion(ctx, userID)
}

func BumpFollowerCacheVersion(ctx context.Context, userID uint) error {
	return Relations.BumpFollowerCacheVersion(ctx, userID)
}

func BumpFriendCacheVersion(ctx context.Context, userID uint) error {
	return Relations.BumpFriendCacheVersion(ctx, userID)
}
