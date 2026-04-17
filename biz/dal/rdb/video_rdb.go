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

const (
	hotVideoCacheTTL    = 1 * time.Minute
	videoDetailCacheTTL = 5 * time.Minute
)

type VideoCache struct {
	client *redis.Client
}

func NewVideoCache(client *redis.Client) VideoCache {
	return VideoCache{client: client}
}

var DefaultVideoCache = NewVideoCache(nil)

func hotVideoCacheVersionKey() string {
	return "video:hot:version"
}

func hotVideoCacheKey(version int64, cursor uint, limit int) string {
	return fmt.Sprintf("video:hot:v%d:%d:%d", version, cursor, limit)
}

func videoDetailCacheKey(videoID uint) string {
	return fmt.Sprintf("video:detail:%d", videoID)
}

func (v VideoCache) redisClient() *redis.Client {
	if v.client != nil {
		return v.client
	}
	return RDB
}

func (v VideoCache) GetHotVideoCacheVersion(ctx context.Context) (int64, error) {
	client := v.redisClient()
	if client == nil {
		return 1, nil
	}

	value, err := client.Get(ctx, hotVideoCacheVersionKey()).Result()
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

func (v VideoCache) GetHotVideoCache(ctx context.Context, version int64, cursor uint, limit int, dest any) (bool, error) {
	client := v.redisClient()
	if client == nil {
		return false, nil
	}

	value, err := client.Get(ctx, hotVideoCacheKey(version, cursor, limit)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return false, err
	}

	return true, nil
}

func (v VideoCache) GetVideoDetailCache(ctx context.Context, videoID uint, dest any) (bool, error) {
	client := v.redisClient()
	if client == nil {
		return false, nil
	}

	value, err := client.Get(ctx, videoDetailCacheKey(videoID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return false, err
	}

	return true, nil
}

func (v VideoCache) SetHotVideoCache(ctx context.Context, version int64, cursor uint, limit int, value any) error {
	client := v.redisClient()
	if client == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return client.Set(ctx, hotVideoCacheKey(version, cursor, limit), payload, hotVideoCacheTTL).Err()
}

func (v VideoCache) SetVideoDetailCache(ctx context.Context, videoID uint, value any) error {
	client := v.redisClient()
	if client == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return client.Set(ctx, videoDetailCacheKey(videoID), payload, videoDetailCacheTTL).Err()
}

func (v VideoCache) BumpHotVideoCacheVersion(ctx context.Context) error {
	client := v.redisClient()
	if client == nil {
		return nil
	}

	return client.Incr(ctx, hotVideoCacheVersionKey()).Err()
}

func (v VideoCache) DeleteVideoDetailCache(ctx context.Context, videoID uint) error {
	client := v.redisClient()
	if client == nil {
		return nil
	}

	return client.Del(ctx, videoDetailCacheKey(videoID)).Err()
}

func GetHotVideoCacheVersion(ctx context.Context) (int64, error) {
	return DefaultVideoCache.GetHotVideoCacheVersion(ctx)
}

func GetHotVideoCache(ctx context.Context, version int64, cursor uint, limit int, dest any) (bool, error) {
	return DefaultVideoCache.GetHotVideoCache(ctx, version, cursor, limit, dest)
}

func GetVideoDetailCache(ctx context.Context, videoID uint, dest any) (bool, error) {
	return DefaultVideoCache.GetVideoDetailCache(ctx, videoID, dest)
}

func SetHotVideoCache(ctx context.Context, version int64, cursor uint, limit int, value any) error {
	return DefaultVideoCache.SetHotVideoCache(ctx, version, cursor, limit, value)
}

func SetVideoDetailCache(ctx context.Context, videoID uint, value any) error {
	return DefaultVideoCache.SetVideoDetailCache(ctx, videoID, value)
}

func BumpHotVideoCacheVersion(ctx context.Context) error {
	return DefaultVideoCache.BumpHotVideoCacheVersion(ctx)
}

func DeleteVideoDetailCache(ctx context.Context, videoID uint) error {
	return DefaultVideoCache.DeleteVideoDetailCache(ctx, videoID)
}
