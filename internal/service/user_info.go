package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-order-lite/internal/dao"
	"go-order-lite/internal/model"
	"go-order-lite/pkg/errno"
	"go-order-lite/pkg/logger"
	"go-order-lite/pkg/redis"
	"time"
)

func GetUserInfo(UserID uint) (*model.User, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d", UserID)

	// 先查 Redis
	val, err := redis.RDB.Get(ctx, cacheKey).Result()
	if err == nil {
		// 空值缓存判断
		if val == "null" {
			logger.Log.Debug("cache hit (empty)")
			return nil, errno.UserNotFound
		}

		var user model.User
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			logger.Log.Debug("cache hit")
			return &user, nil
		}
	}
	logger.Log.Debug("cache miss")
	//  Redis 没有 → 查 MySQL
	user, err := dao.GetUserByID(UserID)
	if err != nil {
		// 空值缓存写入
		redis.RDB.Set(ctx, cacheKey, "null", 30*time.Second)
		return nil, err
	}
	//  写入 Redis
	data, _ := json.Marshal(user)
	redis.RDB.Set(ctx, cacheKey, data, 5*time.Minute)

	return user, nil
}
