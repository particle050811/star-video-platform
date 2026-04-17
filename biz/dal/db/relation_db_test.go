package db

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newMockRelationDB(t *testing.T) (RelationDB, sqlmock.Sqlmock, func()) {
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

	return NewRelationDB(gdb), mock, func() {
		_ = sqlDB.Close()
	}
}

func TestRelationDBFollowUser(t *testing.T) {
	relationDB, mock, cleanup := newMockRelationDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `relations` (`from_user_id`,`to_user_id`,`created_at`) VALUES (?,?,?)")).
		WithArgs(1, 2, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `following_count`=following_count + ?,`updated_at`=? WHERE id = ? AND `users`.`deleted_at` IS NULL")).
		WithArgs(1, sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `follower_count`=follower_count + ?,`updated_at`=? WHERE id = ? AND `users`.`deleted_at` IS NULL")).
		WithArgs(1, sqlmock.AnyArg(), 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := relationDB.FollowUser(context.Background(), 1, 2); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRelationDBUnfollowUserNotFound(t *testing.T) {
	relationDB, mock, cleanup := newMockRelationDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `relations` WHERE from_user_id = ? AND to_user_id = ?")).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	deleted, err := relationDB.UnfollowUser(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if deleted {
		t.Fatal("expected deleted=false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRelationDBListFollowingIDs(t *testing.T) {
	relationDB, mock, cleanup := newMockRelationDB(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	idRows := sqlmock.NewRows([]string{"to_user_id"}).AddRow(3).AddRow(2)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `relations` WHERE from_user_id = ?")).
		WithArgs(1).
		WillReturnRows(countRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `to_user_id` FROM `relations` WHERE from_user_id = ? ORDER BY id DESC LIMIT ?")).
		WithArgs(1, 20).
		WillReturnRows(idRows)

	got, total, err := relationDB.ListFollowingIDs(context.Background(), 1, 0, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(got) != 2 || got[0] != 3 || got[1] != 2 {
		t.Fatalf("unexpected ids: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
