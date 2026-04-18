package video

import (
	"context"
	"errors"
	"testing"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	"video-platform/pkg/parser"
)

type fakeVideoDBStore struct {
	createVideoFn        func(ctx context.Context, video *model.Video) error
	getVideoByIDFn       func(ctx context.Context, videoID uint) (*model.Video, error)
	listVideosByIDsFn    func(ctx context.Context, videoIDs []uint) ([]model.Video, error)
	listVideosByUserIDFn func(ctx context.Context, userID uint, cursor uint, limit int) ([]model.Video, error)
	searchVideosFn       func(ctx context.Context, params db.VideoQuery) ([]model.Video, error)
	listHotVideosFn      func(ctx context.Context, cursor parser.HotVideoCursorValue, limit int) ([]model.Video, error)
}

func (f fakeVideoDBStore) CreateVideo(ctx context.Context, video *model.Video) error {
	return f.createVideoFn(ctx, video)
}

func (f fakeVideoDBStore) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return f.getVideoByIDFn(ctx, videoID)
}

func (f fakeVideoDBStore) ListVideosByIDs(ctx context.Context, videoIDs []uint) ([]model.Video, error) {
	if f.listVideosByIDsFn == nil {
		return nil, nil
	}
	return f.listVideosByIDsFn(ctx, videoIDs)
}

func (f fakeVideoDBStore) ListVideosByUserID(ctx context.Context, userID uint, cursor uint, limit int) ([]model.Video, error) {
	return f.listVideosByUserIDFn(ctx, userID, cursor, limit)
}

func (f fakeVideoDBStore) SearchVideos(ctx context.Context, params db.VideoQuery) ([]model.Video, error) {
	return f.searchVideosFn(ctx, params)
}

func (f fakeVideoDBStore) ListHotVideos(ctx context.Context, cursor parser.HotVideoCursorValue, limit int) ([]model.Video, error) {
	return f.listHotVideosFn(ctx, cursor, limit)
}

type fakeVideoCacheStore struct {
	bumpHotVideoCacheVersionFn func(ctx context.Context) error
	getVideoDetailCacheFn      func(ctx context.Context, videoID uint, dest any) (bool, error)
	setVideoDetailCacheFn      func(ctx context.Context, videoID uint, value any) error
	getHotVideoCacheVersionFn  func(ctx context.Context) (int64, error)
	getHotVideoCacheFn         func(ctx context.Context, version int64, cursor parser.HotVideoCursorValue, limit int, dest any) (bool, error)
	setHotVideoCacheFn         func(ctx context.Context, version int64, cursor parser.HotVideoCursorValue, limit int, value any) error
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

func (f fakeVideoCacheStore) GetHotVideoCache(ctx context.Context, version int64, cursor parser.HotVideoCursorValue, limit int, dest any) (bool, error) {
	return f.getHotVideoCacheFn(ctx, version, cursor, limit, dest)
}

func (f fakeVideoCacheStore) SetHotVideoCache(ctx context.Context, version int64, cursor parser.HotVideoCursorValue, limit int, value any) error {
	return f.setHotVideoCacheFn(ctx, version, cursor, limit, value)
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
				if params.Cursor != 10 || params.Limit != 2 {
					t.Fatalf("unexpected pagination: %+v", params)
				}
				return []model.Video{{ID: 1, Title: "result"}, {ID: 2, Title: "next"}}, nil
			},
		},
	}

	got, err := store.SearchVideos(context.Background(), "go", []uint{1, 2}, 100, 200, "hot", 10, 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].Title != "result" {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.NextCursor != 11 || !got.HasMore {
		t.Fatalf("unexpected cursor result: %+v", got)
	}
}

func TestVideoStoreListHotVideosUsesVersionedCache(t *testing.T) {
	store := videoStore{
		db: fakeVideoDBStore{
			listHotVideosFn: func(ctx context.Context, cursor parser.HotVideoCursorValue, limit int) ([]model.Video, error) {
				t.Fatal("db should not be called on cache hit")
				return nil, nil
			},
		},
		cache: fakeVideoCacheStore{
			getHotVideoCacheVersionFn: func(ctx context.Context) (int64, error) {
				return 3, nil
			},
			getHotVideoCacheFn: func(ctx context.Context, version int64, cursor parser.HotVideoCursorValue, limit int, dest any) (bool, error) {
				wantCursor := parser.HotVideoCursorValue{LikeCount: 10, VisitCount: 20, ID: 30}
				if version != 3 || cursor != wantCursor || limit != 20 {
					t.Fatalf("unexpected cache params version=%d cursor=%+v limit=%d", version, cursor, limit)
				}
				result := dest.(*VideoListResult)
				*result = VideoListResult{Items: []model.Video{{ID: 8, Title: "cached hot"}}}
				return true, nil
			},
		},
	}

	got, err := store.ListHotVideos(context.Background(), parser.HotVideoCursorValue{LikeCount: 10, VisitCount: 20, ID: 30}, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].Title != "cached hot" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestVideoStoreListHotVideosFallsBackToDBAndSetsCache(t *testing.T) {
	var cachedValue VideoListResult

	store := videoStore{
		db: fakeVideoDBStore{
			listHotVideosFn: func(ctx context.Context, cursor parser.HotVideoCursorValue, limit int) ([]model.Video, error) {
				if cursor != (parser.HotVideoCursorValue{}) || limit != 2 {
					t.Fatalf("unexpected db params cursor=%+v limit=%d", cursor, limit)
				}
				return []model.Video{
					{ID: 9, Title: "db hot", LikeCount: 100, VisitCount: 200},
					{ID: 8, Title: "next hot", LikeCount: 90, VisitCount: 300},
				}, nil
			},
		},
		cache: fakeVideoCacheStore{
			getHotVideoCacheVersionFn: func(ctx context.Context) (int64, error) {
				return 7, nil
			},
			getHotVideoCacheFn: func(ctx context.Context, version int64, cursor parser.HotVideoCursorValue, limit int, dest any) (bool, error) {
				return false, nil
			},
			setHotVideoCacheFn: func(ctx context.Context, version int64, cursor parser.HotVideoCursorValue, limit int, value any) error {
				cachedValue = value.(VideoListResult)
				return nil
			},
		},
	}

	got, err := store.ListHotVideos(context.Background(), parser.HotVideoCursorValue{}, 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].Title != "db hot" {
		t.Fatalf("unexpected result: %+v", got)
	}
	if !got.HasMore {
		t.Fatalf("unexpected cursor result: %+v", got)
	}
	wantCursor, err := parser.EncodeHotVideoCursor(parser.HotVideoCursorValue{LikeCount: 100, VisitCount: 200, ID: 9})
	if err != nil {
		t.Fatalf("failed to encode expected cursor: %v", err)
	}
	if got.NextCursorToken != wantCursor {
		t.Fatalf("expected next cursor %q, got %q", wantCursor, got.NextCursorToken)
	}
	if len(cachedValue.Items) != 1 || cachedValue.Items[0].ID != 9 {
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
