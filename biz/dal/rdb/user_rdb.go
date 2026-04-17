package rdb

import (
	"context"
	"fmt"
	"time"
)

const userProfileCacheTTL = 10 * time.Minute

type UserProfileCache struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	AvatarURL      string `json:"avatar_url"`
	FollowingCount int64  `json:"following_count"`
	FollowerCount  int64  `json:"follower_count"`
}

func userProfileCacheKey(userID uint) string {
	return fmt.Sprintf("user:profile:%d", userID)
}

func GetUserProfileCache(ctx context.Context, userID uint) (*UserProfileCache, bool, error) {
	var user UserProfileCache
	ok, err := getJSON(ctx, userProfileCacheKey(userID), &user)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &user, true, nil
}

func SetUserProfileCache(ctx context.Context, userID uint, value any) error {
	return setJSON(ctx, userProfileCacheKey(userID), value, userProfileCacheTTL)
}

func DeleteUserProfileCache(ctx context.Context, userID uint) error {
	return deleteKeys(ctx, userProfileCacheKey(userID))
}
