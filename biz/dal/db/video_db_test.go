package db

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"
	"video-platform/biz/dal/model"
	"video-platform/pkg/parser"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newMockVideoDB(t *testing.T) (VideoDB, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	return NewVideoDB(gdb), mock, func() {
		_ = sqlDB.Close()
	}
}

func TestVideoDBGetVideoByID(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "video_url", "cover_url", "title", "description",
		"visit_count", "like_count", "comment_count", "created_at", "updated_at",
	}).
		AddRow(1, 2, "/static/videos/a.mp4", "/static/video-covers/a.jpg", "title", "desc", 3, 4, 5, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `videos` WHERE `videos`.`id` = ? AND `videos`.`deleted_at` IS NULL ORDER BY `videos`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(rows)

	got, err := videoDB.GetVideoByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Title != "title" {
		t.Fatalf("expected title %q, got %q", "title", got.Title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVideoDBCreateVideo(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `videos` (`user_id`,`video_url`,`cover_url`,`title`,`description`,`visit_count`,`like_count`,`comment_count`,`created_at`,`updated_at`,`deleted_at`) VALUES (?,?,?,?,?,?,?,?,?,?,?)")).
		WithArgs(
			2,
			"/static/videos/a.mp4",
			"/static/video-covers/a.jpg",
			"title",
			"desc",
			0,
			0,
			0,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			nil,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := videoDB.CreateVideo(context.Background(), &model.Video{
		UserID:      2,
		VideoURL:    "/static/videos/a.mp4",
		CoverURL:    "/static/video-covers/a.jpg",
		Title:       "title",
		Description: "desc",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVideoDBListHotVideos(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "video_url", "cover_url", "title", "description",
		"visit_count", "like_count", "comment_count", "created_at", "updated_at",
	}).
		AddRow(1, 2, "/static/videos/a.mp4", "", "hot", "", 10, 9, 8, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `videos` WHERE `videos`.`deleted_at` IS NULL ORDER BY like_count DESC, visit_count DESC, id DESC LIMIT ?")).
		WithArgs(20).
		WillReturnRows(rows)

	got, err := videoDB.ListHotVideos(context.Background(), parser.HotVideoCursorValue{}, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].Title != "hot" {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVideoDBListHotVideosWithCursor(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "video_url", "cover_url", "title", "description",
		"visit_count", "like_count", "comment_count", "created_at", "updated_at",
	}).
		AddRow(7, 2, "/static/videos/a.mp4", "", "hot", "", 20, 10, 8, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `videos` WHERE (like_count < ? OR (like_count = ? AND visit_count < ?) OR (like_count = ? AND visit_count = ? AND id < ?)) AND `videos`.`deleted_at` IS NULL ORDER BY like_count DESC, visit_count DESC, id DESC LIMIT ?")).
		WithArgs(int64(10), int64(10), int64(20), int64(10), int64(20), uint(7), 21).
		WillReturnRows(rows)

	got, err := videoDB.ListHotVideos(context.Background(), parser.HotVideoCursorValue{LikeCount: 10, VisitCount: 20, ID: 7}, 21)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].ID != 7 {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVideoDBListVideosByUserIDUsesSeekCursor(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	anchorTime := time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC)
	anchorRows := sqlmock.NewRows([]string{
		"id", "created_at", "like_count", "visit_count",
	}).AddRow(12, anchorTime, 0, 0)

	resultRows := sqlmock.NewRows([]string{
		"id", "user_id", "video_url", "cover_url", "title", "description",
		"visit_count", "like_count", "comment_count", "created_at", "updated_at",
	}).AddRow(11, 7, "/static/videos/a.mp4", "", "older", "", 1, 2, 3, anchorTime.Add(-time.Minute), anchorTime)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`created_at`,`like_count`,`visit_count` FROM `videos` WHERE id = ? AND user_id = ? AND `videos`.`deleted_at` IS NULL ORDER BY `videos`.`id` LIMIT ?")).
		WithArgs(12, 7, 1).
		WillReturnRows(anchorRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `videos` WHERE user_id = ? AND ((created_at < ?) OR (created_at = ? AND id < ?)) AND `videos`.`deleted_at` IS NULL ORDER BY created_at DESC, id DESC LIMIT ?")).
		WithArgs(7, anchorTime, anchorTime, uint(12), 20).
		WillReturnRows(resultRows)

	got, err := videoDB.ListVideosByUserID(context.Background(), 7, 12, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].ID != 11 {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVideoDBListVideosByUserIDRejectsMissingSeekCursor(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`created_at`,`like_count`,`visit_count` FROM `videos` WHERE id = ? AND user_id = ? AND `videos`.`deleted_at` IS NULL ORDER BY `videos`.`id` LIMIT ?")).
		WithArgs(12, 7, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "like_count", "visit_count"}))

	_, err := videoDB.ListVideosByUserID(context.Background(), 7, 12, 20)
	if !errors.Is(err, ErrVideoCursorInvalid) {
		t.Fatalf("expected error %v, got %v", ErrVideoCursorInvalid, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVideoDBListVideosByUserIDRejectsCursorFromOtherUser(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`created_at`,`like_count`,`visit_count` FROM `videos` WHERE id = ? AND user_id = ? AND `videos`.`deleted_at` IS NULL ORDER BY `videos`.`id` LIMIT ?")).
		WithArgs(12, 7, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "like_count", "visit_count"}))

	_, err := videoDB.ListVideosByUserID(context.Background(), 7, 12, 20)
	if !errors.Is(err, ErrVideoCursorInvalid) {
		t.Fatalf("expected error %v, got %v", ErrVideoCursorInvalid, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVideoDBListVideosByUserIDRejectsSoftDeletedSeekCursor(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`created_at`,`like_count`,`visit_count` FROM `videos` WHERE id = ? AND user_id = ? AND `videos`.`deleted_at` IS NULL ORDER BY `videos`.`id` LIMIT ?")).
		WithArgs(12, 7, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "like_count", "visit_count"}))

	_, err := videoDB.ListVideosByUserID(context.Background(), 7, 12, 20)
	if !errors.Is(err, ErrVideoCursorInvalid) {
		t.Fatalf("expected error %v, got %v", ErrVideoCursorInvalid, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVideoDBSearchVideosUsesSeekCursorForLatestSort(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	anchorTime := time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	anchorRows := sqlmock.NewRows([]string{
		"id", "created_at", "like_count", "visit_count",
	}).AddRow(15, anchorTime, 30, 20)

	resultRows := sqlmock.NewRows([]string{
		"id", "user_id", "video_url", "cover_url", "title", "description",
		"visit_count", "like_count", "comment_count", "created_at", "updated_at",
	}).AddRow(14, 2, "/static/videos/a.mp4", "", "latest", "", 10, 9, 8, anchorTime.Add(-time.Minute), anchorTime)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`created_at`,`like_count`,`visit_count` FROM `videos` WHERE id = ? AND `videos`.`deleted_at` IS NULL ORDER BY `videos`.`id` LIMIT ?")).
		WithArgs(15, 1).
		WillReturnRows(anchorRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `videos` WHERE ((videos.created_at < ?) OR (videos.created_at = ? AND videos.id < ?)) AND `videos`.`deleted_at` IS NULL ORDER BY videos.created_at DESC, videos.id DESC LIMIT ?")).
		WithArgs(anchorTime, anchorTime, uint(15), 10).
		WillReturnRows(resultRows)

	got, err := videoDB.SearchVideos(context.Background(), VideoQuery{
		Cursor: 15,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].ID != 14 {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVideoDBSearchVideosRejectsCursorOutsideSearchFilters(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`created_at`,`like_count`,`visit_count` FROM `videos` WHERE id = ? AND (videos.title LIKE ? ESCAPE '\\\\' OR videos.description LIKE ? ESCAPE '\\\\') AND user_id IN (?) AND videos.created_at >= ? AND videos.created_at <= ? AND `videos`.`deleted_at` IS NULL ORDER BY `videos`.`id` LIMIT ?")).
		WithArgs(15, "%go%", "%go%", 3, time.Unix(100, 0), time.Unix(200, 0), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "like_count", "visit_count"}))

	_, err := videoDB.SearchVideos(context.Background(), VideoQuery{
		Keywords: "go",
		UserIDs:  []uint{3},
		FromDate: 100,
		ToDate:   200,
		Cursor:   15,
		Limit:    10,
	})
	if !errors.Is(err, ErrVideoCursorInvalid) {
		t.Fatalf("expected error %v, got %v", ErrVideoCursorInvalid, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVideoDBSearchVideosEscapesKeywordLikePattern(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "video_url", "cover_url", "title", "description",
		"visit_count", "like_count", "comment_count", "created_at", "updated_at",
	}).AddRow(14, 2, "/static/videos/a.mp4", "", "100%_go", "", 10, 9, 8, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `videos` WHERE (videos.title LIKE ? ESCAPE '\\\\' OR videos.description LIKE ? ESCAPE '\\\\') AND `videos`.`deleted_at` IS NULL ORDER BY videos.created_at DESC, videos.id DESC LIMIT ?")).
		WithArgs("%100\\%\\_go%", "%100\\%\\_go%", 10).
		WillReturnRows(rows)

	got, err := videoDB.SearchVideos(context.Background(), VideoQuery{
		Keywords: "100%_go",
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].ID != 14 {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVideoDBSearchVideosUsesSeekCursorForHotSort(t *testing.T) {
	videoDB, mock, cleanup := newMockVideoDB(t)
	defer cleanup()

	anchorTime := time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	anchorRows := sqlmock.NewRows([]string{
		"id", "created_at", "like_count", "visit_count",
	}).AddRow(15, anchorTime, 30, 20)

	resultRows := sqlmock.NewRows([]string{
		"id", "user_id", "video_url", "cover_url", "title", "description",
		"visit_count", "like_count", "comment_count", "created_at", "updated_at",
	}).AddRow(14, 2, "/static/videos/a.mp4", "", "hot", "", 19, 30, 8, anchorTime.Add(-time.Minute), anchorTime)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`created_at`,`like_count`,`visit_count` FROM `videos` WHERE id = ? AND `videos`.`deleted_at` IS NULL ORDER BY `videos`.`id` LIMIT ?")).
		WithArgs(15, 1).
		WillReturnRows(anchorRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `videos` WHERE ((videos.like_count < ?) OR (videos.like_count = ? AND videos.visit_count < ?) OR (videos.like_count = ? AND videos.visit_count = ? AND videos.id < ?)) AND `videos`.`deleted_at` IS NULL ORDER BY videos.like_count DESC, videos.visit_count DESC, videos.id DESC LIMIT ?")).
		WithArgs(int64(30), int64(30), int64(20), int64(30), int64(20), uint(15), 10).
		WillReturnRows(resultRows)

	got, err := videoDB.SearchVideos(context.Background(), VideoQuery{
		SortBy: "hot",
		Cursor: 15,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].ID != 14 {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
