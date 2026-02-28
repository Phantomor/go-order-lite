package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var (
	RDB *redis.Client
	ctx = context.Background()
)

func InitRedis() error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	if err := RDB.Ping(ctx).Err(); err != nil {
		return err
	}
	log.Println("redis connected successfully")
	return nil
}
