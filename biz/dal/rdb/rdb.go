package rdb

import (
	"context"
	"encoding/json"
	"errors"
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

func getJSON(ctx context.Context, key string, dest any) (bool, error) {
	if RDB == nil {
		return false, nil
	}

	value, err := RDB.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return false, err
	}

	return true, nil
}

func setJSON(ctx context.Context, key string, value any, expiration time.Duration) error {
	if RDB == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return RDB.Set(ctx, key, payload, expiration).Err()
}

func deleteKeys(ctx context.Context, keys ...string) error {
	if RDB == nil || len(keys) == 0 {
		return nil
	}

	return RDB.Del(ctx, keys...).Err()
}

func getCacheVersion(ctx context.Context, key string) (int64, error) {
	if RDB == nil {
		return 1, nil
	}

	value, err := RDB.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 1, nil
		}
		return 0, err
	}

	version, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	if version < 1 {
		return 0, errors.New("cache version must be greater than zero")
	}

	return version, nil
}

func bumpCacheVersion(ctx context.Context, key string) error {
	if RDB == nil {
		return nil
	}

	return RDB.Incr(ctx, key).Err()
}
