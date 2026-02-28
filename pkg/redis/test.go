package redis

import (
	"context"
	"fmt"
	"time"
)

func TestRedis() {
	ctx := context.Background()

	// 1. set key
	err := RDB.Set(ctx, "hello", "world", 10*time.Second).Err()
	if err != nil {
		panic(err)
	}

	// 2. get key
	val, err := RDB.Get(ctx, "hello").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("redis value:", val)
}
