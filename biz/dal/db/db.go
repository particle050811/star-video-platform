package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"video-platform/biz/dal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitMySQL() *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("数据库 DB_DSN 环境变量未设置，请检查 .env 文件")
	}

	createDatabase(dsn)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Warn),
	})

	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	autoMigrate(db)
	DB = db
	return db
}

func createDatabase(dsn string) {
	tmp := strings.Split(dsn, "/")
	baseDSN := tmp[0] + "/"
	dbName := strings.Split(tmp[1], "?")[0]

	// 连接到 MySQL（不指定数据库）
	db, err := sql.Open("mysql", baseDSN)
	if err != nil {
		log.Fatalf("连接 MySQL 失败: %v", err)
	}
	defer db.Close()

	// 创建数据库
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName))
	if err != nil {
		log.Fatalf("创建数据库失败: %v", err)
	}
}

// autoMigrate 自动迁移所有模型
func autoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.User{},
		&model.Relation{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
}
