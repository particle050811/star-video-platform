package interaction

import (
	"context"
	"errors"
	"testing"
	dbdal "video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
)

type fakeInteractionDBStore struct {
	likeVideoFn         func(ctx context.Context, userID, videoID uint) error
	cancelLikeVideoFn   func(ctx context.Context, userID, videoID uint) (bool, error)
	listLikedVideoIDsFn func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.VideoLikeItem, int64, error)
	createCommentFn     func(ctx context.Context, comment *model.Comment) error
	listUserCommentsFn  func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.UserComment, int64, bool, error)
	getCommentByIDFn    func(ctx context.Context, commentID uint) (*model.Comment, error)
	deleteCommentFn     func(ctx context.Context, commentID uint) error
}

func (f fakeInteractionDBStore) LikeVideo(ctx context.Context, userID, videoID uint) error {
	return f.likeVideoFn(ctx, userID, videoID)
}

func (f fakeInteractionDBStore) CancelLikeVideo(ctx context.Context, userID, videoID uint) (bool, error) {
	return f.cancelLikeVideoFn(ctx, userID, videoID)
}

func (f fakeInteractionDBStore) ListLikedVideoIDs(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.VideoLikeItem, int64, error) {
	return f.listLikedVideoIDsFn(ctx, userID, cursor, limit)
}

func (f fakeInteractionDBStore) CreateComment(ctx context.Context, comment *model.Comment) error {
	return f.createCommentFn(ctx, comment)
}

func (f fakeInteractionDBStore) ListUserComments(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.UserComment, int64, bool, error) {
	return f.listUserCommentsFn(ctx, userID, cursor, limit)
}

func (f fakeInteractionDBStore) GetCommentByID(ctx context.Context, commentID uint) (*model.Comment, error) {
	return f.getCommentByIDFn(ctx, commentID)
}

func (f fakeInteractionDBStore) DeleteComment(ctx context.Context, commentID uint) error {
	return f.deleteCommentFn(ctx, commentID)
}

type fakeInteractionVideoStore struct {
	getVideoByIDFn    func(ctx context.Context, videoID uint) (*model.Video, error)
	listVideosByIDsFn func(ctx context.Context, videoIDs []uint) ([]model.Video, error)
}

func (f fakeInteractionVideoStore) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	return f.getVideoByIDFn(ctx, videoID)
}

func (f fakeInteractionVideoStore) ListVideosByIDs(ctx context.Context, videoIDs []uint) ([]model.Video, error) {
	return f.listVideosByIDsFn(ctx, videoIDs)
}

type fakeInteractionCacheStore struct {
	deleteVideoDetailCacheFn       func(ctx context.Context, videoID uint) error
	bumpHotVideoCacheVersionFn     func(ctx context.Context) error
	bumpVideoCommentCacheVersionFn func(ctx context.Context, videoID uint) error
}

func (f fakeInteractionCacheStore) DeleteVideoDetailCache(ctx context.Context, videoID uint) error {
	return f.deleteVideoDetailCacheFn(ctx, videoID)
}

func (f fakeInteractionCacheStore) BumpHotVideoCacheVersion(ctx context.Context) error {
	return f.bumpHotVideoCacheVersionFn(ctx)
}

func (f fakeInteractionCacheStore) BumpVideoCommentCacheVersion(ctx context.Context, videoID uint) error {
	return f.bumpVideoCommentCacheVersionFn(ctx, videoID)
}

func TestInteractionStoreLikeVideoDeletesCaches(t *testing.T) {
	var deletedVideoID uint
	var bumped bool

	store := interactionStore{
		db: fakeInteractionDBStore{
			likeVideoFn: func(ctx context.Context, userID, videoID uint) error {
				if userID != 3 || videoID != 9 {
					t.Fatalf("unexpected like params userID=%d videoID=%d", userID, videoID)
				}
				return nil
			},
		},
		cache: fakeInteractionCacheStore{
			deleteVideoDetailCacheFn: func(ctx context.Context, videoID uint) error {
				deletedVideoID = videoID
				return nil
			},
			bumpHotVideoCacheVersionFn: func(ctx context.Context) error {
				bumped = true
				return nil
			},
			bumpVideoCommentCacheVersionFn: func(ctx context.Context, videoID uint) error {
				t.Fatal("comment cache should not be bumped on like")
				return nil
			},
		},
	}

	if err := store.LikeVideo(context.Background(), 3, 9); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if deletedVideoID != 9 || !bumped {
		t.Fatalf("unexpected cache side effects deleted=%d bumped=%v", deletedVideoID, bumped)
	}
}

func TestInteractionStoreListLikedVideosLoadsVideos(t *testing.T) {
	store := interactionStore{
		db: fakeInteractionDBStore{
			listLikedVideoIDsFn: func(ctx context.Context, userID uint, cursor uint, limit int) ([]dbdal.VideoLikeItem, int64, error) {
				if limit != 2 {
					t.Fatalf("expected limit %d, got %d", 2, limit)
				}
				return []dbdal.VideoLikeItem{
					{ID: 12, VideoID: 101},
					{ID: 11, VideoID: 102},
				}, 5, nil
			},
		},
		videos: fakeInteractionVideoStore{
			listVideosByIDsFn: func(ctx context.Context, videoIDs []uint) ([]model.Video, error) {
				if len(videoIDs) != 1 || videoIDs[0] != 101 {
					t.Fatalf("unexpected video IDs: %+v", videoIDs)
				}
				return []model.Video{{ID: 101, Title: "liked"}}, nil
			},
		},
		cache: fakeInteractionCacheStore{
			deleteVideoDetailCacheFn:       func(ctx context.Context, videoID uint) error { return nil },
			bumpHotVideoCacheVersionFn:     func(ctx context.Context) error { return nil },
			bumpVideoCommentCacheVersionFn: func(ctx context.Context, videoID uint) error { return nil },
		},
	}

	got, err := store.ListLikedVideos(context.Background(), 1, 0, 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].Title != "liked" {
		t.Fatalf("unexpected items: %+v", got.Items)
	}
	if !got.HasMore || got.NextCursor != 12 || got.Total != 5 {
		t.Fatalf("unexpected pagination result: %+v", got)
	}
}

func TestInteractionStoreCreateCommentDeletesCommentCaches(t *testing.T) {
	var deletedVideoID uint
	var bumpedHot bool
	var bumpedCommentVideoID uint

	store := interactionStore{
		db: fakeInteractionDBStore{
			createCommentFn: func(ctx context.Context, comment *model.Comment) error {
				if comment.VideoID != 20 {
					t.Fatalf("unexpected videoID %d", comment.VideoID)
				}
				return nil
			},
		},
		cache: fakeInteractionCacheStore{
			deleteVideoDetailCacheFn: func(ctx context.Context, videoID uint) error {
				deletedVideoID = videoID
				return nil
			},
			bumpHotVideoCacheVersionFn: func(ctx context.Context) error {
				bumpedHot = true
				return nil
			},
			bumpVideoCommentCacheVersionFn: func(ctx context.Context, videoID uint) error {
				bumpedCommentVideoID = videoID
				return nil
			},
		},
	}

	if err := store.CreateComment(context.Background(), &model.Comment{VideoID: 20}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if deletedVideoID != 20 || !bumpedHot || bumpedCommentVideoID != 20 {
		t.Fatalf("unexpected cache side effects deleted=%d hot=%v comment=%d", deletedVideoID, bumpedHot, bumpedCommentVideoID)
	}
}

func TestInteractionStoreDeleteCommentReturnsLookupError(t *testing.T) {
	wantErr := errors.New("lookup failed")
	store := interactionStore{
		db: fakeInteractionDBStore{
			getCommentByIDFn: func(ctx context.Context, commentID uint) (*model.Comment, error) {
				return nil, wantErr
			},
		},
		cache: fakeInteractionCacheStore{
			deleteVideoDetailCacheFn:       func(ctx context.Context, videoID uint) error { return nil },
			bumpHotVideoCacheVersionFn:     func(ctx context.Context) error { return nil },
			bumpVideoCommentCacheVersionFn: func(ctx context.Context, videoID uint) error { return nil },
		},
	}

	if err := store.DeleteComment(context.Background(), 99); !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
