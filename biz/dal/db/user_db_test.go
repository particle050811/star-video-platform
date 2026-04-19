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

func newMockUserDB(t *testing.T) (UserDB, sqlmock.Sqlmock, func()) {
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

	return NewUserDB(gdb), mock, func() {
		_ = sqlDB.Close()
	}
}

func TestUserDBGetUserByID(t *testing.T) {
	userDB, mock, cleanup := newMockUserDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "username", "password", "avatar_url", "following_count", "follower_count"}).
		AddRow(1, "alice", "hashed", "/static/avatars/a.png", 3, 4)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`id` = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(rows)

	got, err := userDB.GetUserByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Username != "alice" {
		t.Fatalf("expected username %q, got %q", "alice", got.Username)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserDBUpdateUserAvatarNotFound(t *testing.T) {
	userDB, mock, cleanup := newMockUserDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `avatar_url`=?,`updated_at`=? WHERE id = ? AND `users`.`deleted_at` IS NULL")).
		WithArgs("/static/avatars/new.png", sqlmock.AnyArg(), 2).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := userDB.UpdateUserAvatar(context.Background(), 2, "/static/avatars/new.png")
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected error %v, got %v", gorm.ErrRecordNotFound, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserDBListUsersByIDs(t *testing.T) {
	userDB, mock, cleanup := newMockUserDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "username", "password", "avatar_url", "following_count", "follower_count"}).
		AddRow(1, "alice", "hashed1", "", 0, 0).
		AddRow(2, "bob", "hashed2", "", 0, 0)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id IN (?,?) AND `users`.`deleted_at` IS NULL")).
		WithArgs(1, 2).
		WillReturnRows(rows)

	got, err := userDB.ListUsersByIDs(context.Background(), []uint{1, 2})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 users, got %d", len(got))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserDBCreateUser(t *testing.T) {
	userDB, mock, cleanup := newMockUserDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users` (`username`,`password`,`avatar_url`,`following_count`,`follower_count`,`created_at`,`updated_at`,`deleted_at`) VALUES (?,?,?,?,?,?,?,?)")).
		WithArgs(
			"alice",
			"hashed",
			"",
			0,
			0,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			nil,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := userDB.CreateUser(context.Background(), &model.User{
		Username: "alice",
		Password: "hashed",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserDBListUserIDsByUsernameEscapesLikePattern(t *testing.T) {
	userDB, mock, cleanup := newMockUserDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(3).
		AddRow(5)

	mock.ExpectQuery("SELECT `id` FROM `users` WHERE username LIKE \\? ESCAPE '\\\\\\\\' AND `users`\\.`deleted_at` IS NULL").
		WithArgs("%a\\%b\\_c\\\\d%").
		WillReturnRows(rows)

	got, err := userDB.ListUserIDsByUsername(context.Background(), "a%b_c\\d")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 2 || got[0] != 3 || got[1] != 5 {
		t.Fatalf("unexpected user ids: %v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
