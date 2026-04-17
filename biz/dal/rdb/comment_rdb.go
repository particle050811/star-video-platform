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

const videoCommentCacheTTL = 1 * time.Minute

type CommentCache struct {
	client *redis.Client
}

func NewCommentCache(client *redis.Client) CommentCache {
	return CommentCache{client: client}
}

var Comments = CommentCache{}

func videoCommentCacheVersionKey(videoID uint) string {
	return fmt.Sprintf("comment:list:version:%d", videoID)
}

func videoCommentCacheKey(videoID uint, version int64, cursor uint, limit int) string {
	return fmt.Sprintf("comment:list:%d:v%d:%d:%d", videoID, version, cursor, limit)
}

func (c CommentCache) redisClient() *redis.Client {
	if c.client != nil {
		return c.client
	}
	return RDB
}

func (c CommentCache) GetVideoCommentCacheVersion(ctx context.Context, videoID uint) (int64, error) {
	client := c.redisClient()
	if client == nil {
		return 1, nil
	}

	value, err := client.Get(ctx, videoCommentCacheVersionKey(videoID)).Result()
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

func (c CommentCache) GetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
	client := c.redisClient()
	if client == nil {
		return false, nil
	}

	value, err := client.Get(ctx, videoCommentCacheKey(videoID, version, cursor, limit)).Result()
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

func (c CommentCache) SetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, value any) error {
	client := c.redisClient()
	if client == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return client.Set(ctx, videoCommentCacheKey(videoID, version, cursor, limit), payload, videoCommentCacheTTL).Err()
}

func (c CommentCache) BumpVideoCommentCacheVersion(ctx context.Context, videoID uint) error {
	client := c.redisClient()
	if client == nil {
		return nil
	}

	return client.Incr(ctx, videoCommentCacheVersionKey(videoID)).Err()
}

func GetVideoCommentCacheVersion(ctx context.Context, videoID uint) (int64, error) {
	return Comments.GetVideoCommentCacheVersion(ctx, videoID)
}

func GetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
	return Comments.GetVideoCommentCache(ctx, videoID, version, cursor, limit, dest)
}

func SetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, value any) error {
	return Comments.SetVideoCommentCache(ctx, videoID, version, cursor, limit, value)
}

func BumpVideoCommentCacheVersion(ctx context.Context, videoID uint) error {
	return Comments.BumpVideoCommentCacheVersion(ctx, videoID)
}
