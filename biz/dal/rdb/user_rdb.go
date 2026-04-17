package rdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const userProfileCacheTTL = 10 * time.Minute

type UserProfileCache struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	AvatarURL      string `json:"avatar_url"`
	FollowingCount int64  `json:"following_count"`
	FollowerCount  int64  `json:"follower_count"`
}

type UserCache struct {
	client *redis.Client
}

func NewUserCache(client *redis.Client) UserCache {
	return UserCache{client: client}
}

var DefaultUserCache = NewUserCache(nil)

func userProfileCacheKey(userID uint) string {
	return fmt.Sprintf("user:profile:%d", userID)
}

func (u UserCache) redisClient() *redis.Client {
	if u.client != nil {
		return u.client
	}
	return RDB
}

func (u UserCache) GetUserProfileCache(ctx context.Context, userID uint) (*UserProfileCache, bool, error) {
	client := u.redisClient()
	if client == nil {
		return nil, false, nil
	}

	var user UserProfileCache
	value, err := client.Get(ctx, userProfileCacheKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}

	if err := json.Unmarshal([]byte(value), &user); err != nil {
		return nil, false, err
	}

	return &user, true, nil
}

func (u UserCache) SetUserProfileCache(ctx context.Context, userID uint, value any) error {
	client := u.redisClient()
	if client == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return client.Set(ctx, userProfileCacheKey(userID), payload, userProfileCacheTTL).Err()
}

func (u UserCache) DeleteUserProfileCache(ctx context.Context, userID uint) error {
	client := u.redisClient()
	if client == nil {
		return nil
	}

	return client.Del(ctx, userProfileCacheKey(userID)).Err()
}

func GetUserProfileCache(ctx context.Context, userID uint) (*UserProfileCache, bool, error) {
	return DefaultUserCache.GetUserProfileCache(ctx, userID)
}

func SetUserProfileCache(ctx context.Context, userID uint, value any) error {
	return DefaultUserCache.SetUserProfileCache(ctx, userID, value)
}

func DeleteUserProfileCache(ctx context.Context, userID uint) error {
	return DefaultUserCache.DeleteUserProfileCache(ctx, userID)
}
