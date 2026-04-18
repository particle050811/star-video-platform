package comment

import (
	"context"
	"errors"
	"testing"
	"time"
	dbdal "video-platform/biz/dal/db"
)

type fakeCommentDBStore struct {
	listVideoCommentsFn func(ctx context.Context, videoID uint, cursor uint, limit int) ([]dbdal.VideoComment, int64, bool, error)
}

func (f fakeCommentDBStore) ListVideoComments(ctx context.Context, videoID uint, cursor uint, limit int) ([]dbdal.VideoComment, int64, bool, error) {
	return f.listVideoCommentsFn(ctx, videoID, cursor, limit)
}

type fakeCommentCacheStore struct {
	getVideoCommentCacheVersionFn  func(ctx context.Context, videoID uint) (int64, error)
	getVideoCommentCacheFn         func(ctx context.Context, videoID uint, version int64, cursor uint, limit int, dest any) (bool, error)
	setVideoCommentCacheFn         func(ctx context.Context, videoID uint, version int64, cursor uint, limit int, value any) error
	bumpVideoCommentCacheVersionFn func(ctx context.Context, videoID uint) error
}

func (f fakeCommentCacheStore) GetVideoCommentCacheVersion(ctx context.Context, videoID uint) (int64, error) {
	return f.getVideoCommentCacheVersionFn(ctx, videoID)
}

func (f fakeCommentCacheStore) GetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
	return f.getVideoCommentCacheFn(ctx, videoID, version, cursor, limit, dest)
}

func (f fakeCommentCacheStore) SetVideoCommentCache(ctx context.Context, videoID uint, version int64, cursor uint, limit int, value any) error {
	return f.setVideoCommentCacheFn(ctx, videoID, version, cursor, limit, value)
}

func (f fakeCommentCacheStore) BumpVideoCommentCacheVersion(ctx context.Context, videoID uint) error {
	return f.bumpVideoCommentCacheVersionFn(ctx, videoID)
}

func TestCommentStoreListVideoCommentsUsesCache(t *testing.T) {
	store := commentStore{
		db: fakeCommentDBStore{
			listVideoCommentsFn: func(ctx context.Context, videoID uint, cursor uint, limit int) ([]dbdal.VideoComment, int64, bool, error) {
				t.Fatal("db should not be called on cache hit")
				return nil, 0, false, nil
			},
		},
		cache: fakeCommentCacheStore{
			getVideoCommentCacheVersionFn: func(ctx context.Context, videoID uint) (int64, error) {
				return 3, nil
			},
			getVideoCommentCacheFn: func(ctx context.Context, videoID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
				payload := dest.(*videoCommentCachePayload)
				*payload = videoCommentCachePayload{
					Items:      []VideoComment{{ID: 10, UserID: 8, Content: "cached"}},
					Total:      5,
					NextCursor: 10,
					HasMore:    true,
				}
				return true, nil
			},
			setVideoCommentCacheFn: func(ctx context.Context, videoID uint, version int64, cursor uint, limit int, value any) error {
				t.Fatal("cache set should not be called on cache hit")
				return nil
			},
			bumpVideoCommentCacheVersionFn: func(ctx context.Context, videoID uint) error { return nil },
		},
	}

	got, err := store.ListVideoComments(context.Background(), 1, 0, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].Content != "cached" {
		t.Fatalf("unexpected items: %+v", got.Items)
	}
	if got.Total != 5 || got.NextCursor != 10 || !got.HasMore {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestCommentStoreListVideoCommentsFallsBackToDBAndSetsCache(t *testing.T) {
	var cached videoCommentCachePayload
	now := time.Now()

	store := commentStore{
		db: fakeCommentDBStore{
			listVideoCommentsFn: func(ctx context.Context, videoID uint, cursor uint, limit int) ([]dbdal.VideoComment, int64, bool, error) {
				return []dbdal.VideoComment{
					{ID: 9, UserID: 2, Content: "first", LikeCount: 7, CreatedAt: now},
					{ID: 8, UserID: 3, Content: "second", LikeCount: 1, CreatedAt: now},
				}, 12, true, nil
			},
		},
		cache: fakeCommentCacheStore{
			getVideoCommentCacheVersionFn: func(ctx context.Context, videoID uint) (int64, error) {
				return 4, nil
			},
			getVideoCommentCacheFn: func(ctx context.Context, videoID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
				return false, nil
			},
			setVideoCommentCacheFn: func(ctx context.Context, videoID uint, version int64, cursor uint, limit int, value any) error {
				cached = value.(videoCommentCachePayload)
				return nil
			},
			bumpVideoCommentCacheVersionFn: func(ctx context.Context, videoID uint) error { return nil },
		},
	}

	got, err := store.ListVideoComments(context.Background(), 2, 0, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Items) != 2 || got.Items[0].Content != "first" {
		t.Fatalf("unexpected items: %+v", got.Items)
	}
	if got.NextCursor != 8 || !got.HasMore || got.Total != 12 {
		t.Fatalf("unexpected result: %+v", got)
	}
	if len(cached.Items) != 2 || cached.NextCursor != 8 || cached.Total != 12 {
		t.Fatalf("unexpected cached payload: %+v", cached)
	}
}

func TestCommentStoreListVideoCommentsReturnsVersionError(t *testing.T) {
	wantErr := errors.New("version error")
	store := commentStore{
		db: fakeCommentDBStore{
			listVideoCommentsFn: func(ctx context.Context, videoID uint, cursor uint, limit int) ([]dbdal.VideoComment, int64, bool, error) {
				return nil, 0, false, nil
			},
		},
		cache: fakeCommentCacheStore{
			getVideoCommentCacheVersionFn: func(ctx context.Context, videoID uint) (int64, error) {
				return 0, wantErr
			},
			getVideoCommentCacheFn: func(ctx context.Context, videoID uint, version int64, cursor uint, limit int, dest any) (bool, error) {
				return false, nil
			},
			setVideoCommentCacheFn: func(ctx context.Context, videoID uint, version int64, cursor uint, limit int, value any) error {
				return nil
			},
			bumpVideoCommentCacheVersionFn: func(ctx context.Context, videoID uint) error { return nil },
		},
	}

	if _, err := store.ListVideoComments(context.Background(), 1, 0, 20); !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
