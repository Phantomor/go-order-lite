package service

import (
	"context"
	"go-order-lite/pkg/logger"
	"go-order-lite/pkg/redis"
	"strconv"
	"time"

	redisV9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const OrderDelayQueueKey = "order:delay_queue"

// PushOrderToDelayQueue 将订单加入延迟队列
func PushOrderToDelayQueue(ctx context.Context, orderID uint, delay time.Duration) error {
	score := float64(time.Now().Add(delay).Unix())
	return redis.RDB.ZAdd(ctx, OrderDelayQueueKey, redisV9.Z{
		Score:  score,
		Member: strconv.Itoa(int(orderID)),
	}).Err()
}

// StartDelayQueueConsumer 启动延迟队列消费者
func StartDelayQueueConsumer(ctx context.Context) {
	ticker := time.NewTicker(time.Second) // 每 1 秒轮询一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("delay queue consumer stopped")
			return
		case <-ticker.C:
			processExpiredOrders(ctx)
		}
	}
}

func processExpiredOrders(ctx context.Context) {
	now := time.Now().Unix()

	// 1. 查出所有到期的 orderID (Score <= 当前时间戳)
	opt := &redisV9.ZRangeBy{
		Min: "-inf",
		Max: strconv.FormatInt(now, 10),
	}
	members, err := redis.RDB.ZRangeByScore(ctx, OrderDelayQueueKey, opt).Result()
	if err != nil {
		logger.Log.Error("zrangebyscore failed", zap.Error(err))
		return
	}

	// 2. 遍历处理
	for _, member := range members {
		// 3. 尝试从 ZSet 中删除该订单，删除成功代表抢到了该任务（防止多实例并发重复取消）
		removed, err := redis.RDB.ZRem(ctx, OrderDelayQueueKey, member).Result()
		if err != nil {
			logger.Log.Error("zrem failed", zap.Error(err))
			continue
		}

		// 抢占成功，执行取消逻辑
		if removed == 1 {
			orderID, _ := strconv.Atoi(member)
			err := SystemCancelOrder(uint(orderID))
			if err != nil {
				logger.Log.Error("system cancel order failed", zap.Uint("orderID", uint(orderID)), zap.Error(err))
			} else {
				logger.Log.Info("order auto canceled via delay queue", zap.Uint("orderID", uint(orderID)))
			}
		}
	}
}
