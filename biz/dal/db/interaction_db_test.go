package db

import (
	"context"
	"regexp"
	"testing"
	"video-platform/biz/dal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newMockInteractionDB(t *testing.T) (InteractionDB, sqlmock.Sqlmock, func()) {
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

	return NewInteractionDB(gdb), mock, func() {
		_ = sqlDB.Close()
	}
}

func TestInteractionDBLikeVideoReturnsNotFoundWhenVideoMissing(t *testing.T) {
	interactionDB, mock, cleanup := newMockInteractionDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `video_likes` (`user_id`,`video_id`,`created_at`) VALUES (?,?,?)")).
		WithArgs(1, 2, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `videos` SET `like_count`=like_count + ?,`updated_at`=? WHERE id = ? AND `videos`.`deleted_at` IS NULL")).
		WithArgs(1, sqlmock.AnyArg(), 2).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := interactionDB.LikeVideo(context.Background(), 1, 2)
	if err == nil || !IsRecordNotFound(err) {
		t.Fatalf("expected record not found error, got %v", err)
	}
}

func TestInteractionDBCancelLikeVideoReturnsNotFoundWhenVideoMissing(t *testing.T) {
	interactionDB, mock, cleanup := newMockInteractionDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `video_likes` WHERE user_id = ? AND video_id = ?")).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `videos` SET `like_count`=like_count - ?,`updated_at`=? WHERE id = ? AND `videos`.`deleted_at` IS NULL")).
		WithArgs(1, sqlmock.AnyArg(), 2).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	deleted, err := interactionDB.CancelLikeVideo(context.Background(), 1, 2)
	if deleted {
		t.Fatal("expected deleted=false on rollback")
	}
	if err == nil || !IsRecordNotFound(err) {
		t.Fatalf("expected record not found error, got %v", err)
	}
}

func TestInteractionDBCreateCommentReturnsNotFoundWhenVideoMissing(t *testing.T) {
	interactionDB, mock, cleanup := newMockInteractionDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `comments` (`video_id`,`user_id`,`content`,`like_count`,`created_at`,`updated_at`,`deleted_at`) VALUES (?,?,?,?,?,?,?)")).
		WithArgs(2, 1, "hi", 0, sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `videos` SET `comment_count`=comment_count + ?,`updated_at`=? WHERE id = ? AND `videos`.`deleted_at` IS NULL")).
		WithArgs(1, sqlmock.AnyArg(), 2).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := interactionDB.CreateComment(context.Background(), &model.Comment{
		UserID:  1,
		VideoID: 2,
		Content: "hi",
	})
	if err == nil || !IsRecordNotFound(err) {
		t.Fatalf("expected record not found error, got %v", err)
	}
}

func TestInteractionDBDeleteCommentReturnsNotFoundWhenDeleteMissesAfterRead(t *testing.T) {
	interactionDB, mock, cleanup := newMockInteractionDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "video_id", "user_id", "content", "like_count", "created_at", "updated_at", "deleted_at"}).
		AddRow(3, 2, 1, "hi", 0, nil, nil, nil)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `comments` WHERE `comments`.`id` = ? AND `comments`.`deleted_at` IS NULL ORDER BY `comments`.`id` LIMIT ?")).
		WithArgs(3, 1).
		WillReturnRows(rows)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `comments` SET `deleted_at`=? WHERE `comments`.`id` = ? AND `comments`.`deleted_at` IS NULL")).
		WithArgs(sqlmock.AnyArg(), 3).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := interactionDB.DeleteComment(context.Background(), 3)
	if err == nil || !IsRecordNotFound(err) {
		t.Fatalf("expected record not found error, got %v", err)
	}
}
