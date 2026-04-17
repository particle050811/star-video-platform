package rdb

import (
	"context"
	"fmt"
	"time"
)

const videoCommentCacheTTL = 1 * time.Minute

func videoCommentCacheKey(videoID uint, offset, limit int) string {
	return fmt.Sprintf("comment:list:%d:%d:%d", videoID, offset, limit)
}

func GetVideoCommentCache(ctx context.Context, videoID uint, offset, limit int, dest any) (bool, error) {
	return getJSON(ctx, videoCommentCacheKey(videoID, offset, limit), dest)
}

func SetVideoCommentCache(ctx context.Context, videoID uint, offset, limit int, value any) error {
	return setJSON(ctx, videoCommentCacheKey(videoID, offset, limit), value, videoCommentCacheTTL)
}

func DeleteVideoCommentCaches(ctx context.Context, videoID uint) error {
	return deleteByPattern(ctx, fmt.Sprintf("comment:list:%d:*", videoID))
}
