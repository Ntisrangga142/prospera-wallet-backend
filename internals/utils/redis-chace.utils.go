package utils

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func CacheHit(rctx context.Context, rdb *redis.Client, redisKey string, data any) error {
	cmd := rdb.Get(rctx, redisKey)
	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil {
			log.Printf("Key %s does not exist.\n", redisKey)
		} else {
			log.Printf("Redis Error.\nCause: %s\n", cmd.Err())
		}
		return cmd.Err()
	}

	cmdByte, err := cmd.Bytes()
	if err != nil {
		log.Println("Internal Server Error.\nCause:", err)
		return err
	}

	if err := json.Unmarshal(cmdByte, data); err != nil {
		log.Println("Internal Server Error.\nCause:", err)
		return err
	}

	return nil
}

func RenewCache(rctx context.Context, rdb *redis.Client, redisKey string, data any, waktu time.Duration) error {
	bt, err := json.Marshal(data)
	if err != nil {
		log.Println("Internal Server Error.\nCause:", err)
		return err
	}

	if err := rdb.Set(rctx, redisKey, bt, waktu*time.Minute).Err(); err != nil {
		log.Printf("Redis Error.\nCause: %s\n", err)
		return err
	}

	return nil
}

func InvalidateCache(rctx context.Context, rdb *redis.Client, redisKey string) error {
	if err := rdb.Del(rctx, redisKey).Err(); err != nil {
		log.Printf("Redis Error.\nCause: %s\n", err)
		return err
	}
	return nil
}
