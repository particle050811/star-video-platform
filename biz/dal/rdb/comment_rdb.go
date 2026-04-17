package rdb

import (
	"context"
	"fmt"
	"time"
)

const videoCommentCacheTTL = 1 * time.Minute

func videoCommentCacheVersionKey(videoID uint) string {
	return fmt.Sprintf("comment:list:version:%d", videoID)
}

func videoCommentCacheKey(videoID uint, version int64, cursor uint, limit int) string {
	return fmt.Sprintf("comment:list:%d:v%d:%d:%d", videoID, version, cursor, limit)
}

func GetVideoCommentCacheVersion(ctx context.Context, videoID uint) (int64, error) {
	return getCacheVersion(ctx, videoCommentCacheVersionKey(videoID))
}

func GetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
	return getJSON(ctx, videoCommentCacheKey(videoID, version, cursor, limit), dest)
}

func SetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, value any) error {
	return setJSON(ctx, videoCommentCacheKey(videoID, version, cursor, limit), value, videoCommentCacheTTL)
}

func BumpVideoCommentCacheVersion(ctx context.Context, videoID uint) error {
	return bumpCacheVersion(ctx, videoCommentCacheVersionKey(videoID))
}
