package rdb

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		log.Fatal("REDIS REDIS_ADDR 环境变量未设置，请检查 .env 文件")
	}

	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		parsed, err := strconv.Atoi(dbStr)
		if err != nil {
			log.Fatalf("REDIS_DB=%q 非法，服务启动失败", dbStr)
		}
		db = parsed
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		log.Fatalf("[Redis] 连接失败，服务启动失败 addr=%s: %v", addr, err)
	}

	RDB = rdb
	return rdb
}
