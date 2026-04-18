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

const chatMessageCacheTTL = 1 * time.Minute

type ChatCache struct {
	client *redis.Client
}

func NewChatCache(client *redis.Client) ChatCache {
	return ChatCache{client: client}
}

var Chats = ChatCache{}

func chatMessageCacheVersionKey(roomID uint) string {
	return fmt.Sprintf("chat:messages:version:%d", roomID)
}

func chatMessageCacheKey(roomID uint, version int64, cursor uint, limit int) string {
	return fmt.Sprintf("chat:messages:%d:v%d:%d:%d", roomID, version, cursor, limit)
}

func (c ChatCache) redisClient() *redis.Client {
	if c.client != nil {
		return c.client
	}
	return RDB
}

func (c ChatCache) GetChatMessageCacheVersion(ctx context.Context, roomID uint) (int64, error) {
	client := c.redisClient()
	if client == nil {
		return 1, nil
	}

	value, err := client.Get(ctx, chatMessageCacheVersionKey(roomID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}

	version, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	if version < 0 {
		return 0, errors.New("cache version must not be negative")
	}

	return version, nil
}

func (c ChatCache) GetChatMessageCache(ctx context.Context, roomID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
	client := c.redisClient()
	if client == nil {
		return false, nil
	}

	value, err := client.Get(ctx, chatMessageCacheKey(roomID, version, cursor, limit)).Result()
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

func (c ChatCache) SetChatMessageCache(ctx context.Context, roomID uint, version int64, cursor uint, limit int, value any) error {
	client := c.redisClient()
	if client == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return client.Set(ctx, chatMessageCacheKey(roomID, version, cursor, limit), payload, chatMessageCacheTTL).Err()
}

func (c ChatCache) BumpChatMessageCacheVersion(ctx context.Context, roomID uint) error {
	client := c.redisClient()
	if client == nil {
		return nil
	}

	return client.Incr(ctx, chatMessageCacheVersionKey(roomID)).Err()
}

func GetChatMessageCacheVersion(ctx context.Context, roomID uint) (int64, error) {
	return Chats.GetChatMessageCacheVersion(ctx, roomID)
}

func GetChatMessageCache(ctx context.Context, roomID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
	return Chats.GetChatMessageCache(ctx, roomID, version, cursor, limit, dest)
}

func SetChatMessageCache(ctx context.Context, roomID uint, version int64, cursor uint, limit int, value any) error {
	return Chats.SetChatMessageCache(ctx, roomID, version, cursor, limit, value)
}

func BumpChatMessageCacheVersion(ctx context.Context, roomID uint) error {
	return Chats.BumpChatMessageCacheVersion(ctx, roomID)
}
