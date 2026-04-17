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

func hotVideoCacheVersionKey() string {
	return "video:hot:version"
}

func hotVideoCacheKey(version int64, offset, limit int) string {
	return fmt.Sprintf("video:hot:v%d:%d:%d", version, offset, limit)
}

func GetHotVideoCacheVersion(ctx context.Context) (int64, error) {
	return getCacheVersion(ctx, hotVideoCacheVersionKey())
}

func videoDetailCacheKey(videoID uint) string {
	return fmt.Sprintf("video:detail:%d", videoID)
}

func GetHotVideoCache(ctx context.Context, version int64, offset, limit int, dest any) (bool, error) {
	return getJSON(ctx, hotVideoCacheKey(version, offset, limit), dest)
}

func GetVideoDetailCache(ctx context.Context, videoID uint, dest any) (bool, error) {
	return getJSON(ctx, videoDetailCacheKey(videoID), dest)
}

func SetHotVideoCache(ctx context.Context, version int64, offset, limit int, value any) error {
	return setJSON(ctx, hotVideoCacheKey(version, offset, limit), value, hotVideoCacheTTL)
}

func SetVideoDetailCache(ctx context.Context, videoID uint, value any) error {
	return setJSON(ctx, videoDetailCacheKey(videoID), value, videoDetailCacheTTL)
}

func BumpHotVideoCacheVersion(ctx context.Context) error {
	return bumpCacheVersion(ctx, hotVideoCacheVersionKey())
}

func DeleteVideoDetailCache(ctx context.Context, videoID uint) error {
	return deleteKeys(ctx, videoDetailCacheKey(videoID))
}
