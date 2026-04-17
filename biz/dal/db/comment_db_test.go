package db

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newMockCommentDB(t *testing.T) (CommentDB, sqlmock.Sqlmock, func()) {
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

	return NewCommentDB(gdb), mock, func() {
		_ = sqlDB.Close()
	}
}

func TestCommentDBListVideoComments(t *testing.T) {
	commentDB, mock, cleanup := newMockCommentDB(t)
	defer cleanup()

	now := time.Now()
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
	commentRows := sqlmock.NewRows([]string{"id", "user_id", "content", "like_count", "created_at"}).
		AddRow(4, 11, "first", 2, now).
		AddRow(3, 12, "second", 1, now).
		AddRow(2, 13, "extra", 0, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `comments` WHERE video_id = ? AND `comments`.`deleted_at` IS NULL")).
		WithArgs(9).
		WillReturnRows(countRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, content, like_count, created_at FROM `comments` WHERE video_id = ? AND id < ? AND `comments`.`deleted_at` IS NULL ORDER BY id DESC LIMIT ?")).
		WithArgs(9, 5, 3).
		WillReturnRows(commentRows)

	got, total, hasMore, err := commentDB.ListVideoComments(context.Background(), 9, 5, 2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	if !hasMore {
		t.Fatal("expected hasMore=true")
	}
	if len(got) != 2 || got[0].ID != 4 || got[1].ID != 3 {
		t.Fatalf("unexpected comments: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
