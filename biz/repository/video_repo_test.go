package repository

import (
	"context"
	"errors"
	"testing"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
)

type fakeVideoDBStore struct {
	createVideoFn        func(ctx context.Context, video *model.Video) error
	getVideoByIDFn       func(ctx context.Context, videoID uint) (*model.Video, error)
	listVideosByUserIDFn func(ctx context.Context, userID uint, offset, limit int) ([]model.Video, error)
	searchVideosFn       func(ctx context.Context, params db.VideoQuery) ([]model.Video, error)
	listHotVideosFn      func(ctx context.Context, offset, limit int) ([]model.Video, error)
}

func (f fakeVideoDBStore) CreateVideo(ctx context.Context, video *model.Video) error {
	return f.createVideoFn(ctx, video)
}

func (f fakeVideoDBStore) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return f.getVideoByIDFn(ctx, videoID)
}

func (f fakeVideoDBStore) ListVideosByUserID(ctx context.Context, userID uint, offset, limit int) ([]model.Video, error) {
	return f.listVideosByUserIDFn(ctx, userID, offset, limit)
}

func (f fakeVideoDBStore) SearchVideos(ctx context.Context, params db.VideoQuery) ([]model.Video, error) {
	return f.searchVideosFn(ctx, params)
}

func (f fakeVideoDBStore) ListHotVideos(ctx context.Context, offset, limit int) ([]model.Video, error) {
	return f.listHotVideosFn(ctx, offset, limit)
}

type fakeVideoCacheStore struct {
	bumpHotVideoCacheVersionFn func(ctx context.Context) error
	getVideoDetailCacheFn      func(ctx context.Context, videoID uint, dest any) (bool, error)
	setVideoDetailCacheFn      func(ctx context.Context, videoID uint, value any) error
	getHotVideoCacheVersionFn  func(ctx context.Context) (int64, error)
	getHotVideoCacheFn         func(ctx context.Context, version int64, offset, limit int, dest any) (bool, error)
	setHotVideoCacheFn         func(ctx context.Context, version int64, offset, limit int, value any) error
}

func (f fakeVideoCacheStore) BumpHotVideoCacheVersion(ctx context.Context) error {
	return f.bumpHotVideoCacheVersionFn(ctx)
}

func (f fakeVideoCacheStore) GetVideoDetailCache(ctx context.Context, videoID uint, dest any) (bool, error) {
	return f.getVideoDetailCacheFn(ctx, videoID, dest)
}

func (f fakeVideoCacheStore) SetVideoDetailCache(ctx context.Context, videoID uint, value any) error {
	return f.setVideoDetailCacheFn(ctx, videoID, value)
}

func (f fakeVideoCacheStore) GetHotVideoCacheVersion(ctx context.Context) (int64, error) {
	return f.getHotVideoCacheVersionFn(ctx)
}

func (f fakeVideoCacheStore) GetHotVideoCache(ctx context.Context, version int64, offset, limit int, dest any) (bool, error) {
	return f.getHotVideoCacheFn(ctx, version, offset, limit, dest)
}

func (f fakeVideoCacheStore) SetHotVideoCache(ctx context.Context, version int64, offset, limit int, value any) error {
	return f.setHotVideoCacheFn(ctx, version, offset, limit, value)
}

func TestVideoStoreCreateVideoBumpsCacheVersion(t *testing.T) {
	var bumped bool

	store := videoStore{
		db: fakeVideoDBStore{
			createVideoFn: func(ctx context.Context, video *model.Video) error {
				if video.Title != "title" {
					t.Fatalf("expected title %q, got %q", "title", video.Title)
				}
				return nil
			},
		},
		cache: fakeVideoCacheStore{
			bumpHotVideoCacheVersionFn: func(ctx context.Context) error {
				bumped = true
				return nil
			},
		},
	}

	if err := store.CreateVideo(context.Background(), &model.Video{Title: "title"}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !bumped {
		t.Fatal("expected hot video cache version to be bumped")
	}
}

func TestVideoStoreGetVideoByIDUsesCache(t *testing.T) {
	store := videoStore{
		db: fakeVideoDBStore{
			getVideoByIDFn: func(ctx context.Context, videoID uint) (*model.Video, error) {
				t.Fatal("db should not be called on cache hit")
				return nil, nil
			},
		},
		cache: fakeVideoCacheStore{
			getVideoDetailCacheFn: func(ctx context.Context, videoID uint, dest any) (bool, error) {
				video := dest.(*model.Video)
				video.ID = videoID
				video.Title = "cached"
				return true, nil
			},
			setVideoDetailCacheFn: func(ctx context.Context, videoID uint, value any) error {
				t.Fatal("cache set should not be called on cache hit")
				return nil
			},
		},
	}

	got, err := store.GetVideoByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Title != "cached" {
		t.Fatalf("expected title %q, got %q", "cached", got.Title)
	}
}

func TestVideoStoreGetVideoByIDFallsBackToDBAndSetsCache(t *testing.T) {
	var cached *model.Video

	store := videoStore{
		db: fakeVideoDBStore{
			getVideoByIDFn: func(ctx context.Context, videoID uint) (*model.Video, error) {
				return &model.Video{ID: videoID, Title: "db"}, nil
			},
		},
		cache: fakeVideoCacheStore{
			getVideoDetailCacheFn: func(ctx context.Context, videoID uint, dest any) (bool, error) {
				return false, nil
			},
			setVideoDetailCacheFn: func(ctx context.Context, videoID uint, value any) error {
				video := value.(*model.Video)
				cached = video
				return nil
			},
		},
	}

	got, err := store.GetVideoByID(context.Background(), 6)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Title != "db" {
		t.Fatalf("expected title %q, got %q", "db", got.Title)
	}
	if cached == nil || cached.ID != 6 {
		t.Fatalf("unexpected cached value: %+v", cached)
	}
}

func TestVideoStoreSearchVideosBuildsQuery(t *testing.T) {
	store := videoStore{
		db: fakeVideoDBStore{
			searchVideosFn: func(ctx context.Context, params db.VideoQuery) ([]model.Video, error) {
				if params.Keywords != "go" || params.SortBy != "hot" {
					t.Fatalf("unexpected params: %+v", params)
				}
				if len(params.UserIDs) != 2 || params.UserIDs[0] != 1 || params.UserIDs[1] != 2 {
					t.Fatalf("unexpected userIDs: %+v", params.UserIDs)
				}
				if params.Offset != 10 || params.Limit != 20 {
					t.Fatalf("unexpected pagination: %+v", params)
				}
				return []model.Video{{ID: 1, Title: "result"}}, nil
			},
		},
	}

	got, err := store.SearchVideos(context.Background(), "go", []uint{1, 2}, 100, 200, "hot", 10, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].Title != "result" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestVideoStoreListHotVideosUsesVersionedCache(t *testing.T) {
	store := videoStore{
		db: fakeVideoDBStore{
			listHotVideosFn: func(ctx context.Context, offset, limit int) ([]model.Video, error) {
				t.Fatal("db should not be called on cache hit")
				return nil, nil
			},
		},
		cache: fakeVideoCacheStore{
			getHotVideoCacheVersionFn: func(ctx context.Context) (int64, error) {
				return 3, nil
			},
			getHotVideoCacheFn: func(ctx context.Context, version int64, offset, limit int, dest any) (bool, error) {
				if version != 3 || offset != 10 || limit != 20 {
					t.Fatalf("unexpected cache params version=%d offset=%d limit=%d", version, offset, limit)
				}
				videos := dest.(*[]model.Video)
				*videos = []model.Video{{ID: 8, Title: "cached hot"}}
				return true, nil
			},
		},
	}

	got, err := store.ListHotVideos(context.Background(), 10, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].Title != "cached hot" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestVideoStoreListHotVideosFallsBackToDBAndSetsCache(t *testing.T) {
	var cachedValue []model.Video

	store := videoStore{
		db: fakeVideoDBStore{
			listHotVideosFn: func(ctx context.Context, offset, limit int) ([]model.Video, error) {
				return []model.Video{{ID: 9, Title: "db hot"}}, nil
			},
		},
		cache: fakeVideoCacheStore{
			getHotVideoCacheVersionFn: func(ctx context.Context) (int64, error) {
				return 7, nil
			},
			getHotVideoCacheFn: func(ctx context.Context, version int64, offset, limit int, dest any) (bool, error) {
				return false, nil
			},
			setHotVideoCacheFn: func(ctx context.Context, version int64, offset, limit int, value any) error {
				cachedValue = value.([]model.Video)
				return nil
			},
		},
	}

	got, err := store.ListHotVideos(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].Title != "db hot" {
		t.Fatalf("unexpected result: %+v", got)
	}
	if len(cachedValue) != 1 || cachedValue[0].ID != 9 {
		t.Fatalf("unexpected cached value: %+v", cachedValue)
	}
}

func TestVideoStoreGetVideoByIDReturnsDBError(t *testing.T) {
	wantErr := errors.New("db error")
	store := videoStore{
		db: fakeVideoDBStore{
			getVideoByIDFn: func(ctx context.Context, videoID uint) (*model.Video, error) {
				return nil, wantErr
			},
		},
		cache: fakeVideoCacheStore{
			getVideoDetailCacheFn: func(ctx context.Context, videoID uint, dest any) (bool, error) {
				return false, nil
			},
		},
	}

	_, err := store.GetVideoByID(context.Background(), 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
