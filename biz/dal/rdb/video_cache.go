package rdb

import (
	"context"
	"fmt"
	"time"
)

const (
	hotVideoCacheTTL    = 1 * time.Minute
	videoDetailCacheTTL = 5 * time.Minute
)

func hotVideoCacheKey(offset, limit int) string {
	return fmt.Sprintf("video:hot:%d:%d", offset, limit)
}

func videoDetailCacheKey(videoID uint) string {
	return fmt.Sprintf("video:detail:%d", videoID)
}

func GetHotVideoCache(ctx context.Context, offset, limit int, dest any) (bool, error) {
	return getJSON(ctx, hotVideoCacheKey(offset, limit), dest)
}

func GetVideoDetailCache(ctx context.Context, videoID uint, dest any) (bool, error) {
	return getJSON(ctx, videoDetailCacheKey(videoID), dest)
}

func SetHotVideoCache(ctx context.Context, offset, limit int, value any) error {
	return setJSON(ctx, hotVideoCacheKey(offset, limit), value, hotVideoCacheTTL)
}

func SetVideoDetailCache(ctx context.Context, videoID uint, value any) error {
	return setJSON(ctx, videoDetailCacheKey(videoID), value, videoDetailCacheTTL)
}

func DeleteHotVideoCaches(ctx context.Context) error {
	return deleteByPattern(ctx, "video:hot:*")
}

func DeleteVideoDetailCache(ctx context.Context, videoID uint) error {
	return deleteKeys(ctx, videoDetailCacheKey(videoID))
}
