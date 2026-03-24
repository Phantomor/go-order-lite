package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"go-order-lite/pkg/config"
	"go-order-lite/pkg/logger"
	"go-order-lite/pkg/redis"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

var PointConsumer rocketmq.PushConsumer

// InitPointConsumer 初始化并启动积分发放的消费者
func InitPointConsumer() error {
	var err error
	PointConsumer, err = rocketmq.NewPushConsumer(
		consumer.WithNameServer(config.Cfg.RocketMQ.NameServers),
		consumer.WithMaxReconsumeTimes(3),
		consumer.WithGroupName("reward-point-group"),
	)
	if err != nil {
		return err
	}
	err = PointConsumer.Subscribe("ORDER_PAID", consumer.MessageSelector{},
		func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, msg := range msgs {
				var event map[string]interface{}
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					logger.Log.Error("invalid message body", zap.String("body", string(msg.Body)), zap.Error(err))
					continue // 数据格式错误，没必要重试，直接跳过
				}
				// 模拟业务逻辑：为用户增加积分
				orderID := uint(event["order_id"].(float64))
				userID := uint(event["user_id"].(float64))
				amount := event["amount"].(float64)

				// 幂等性校验
				lockKey := fmt.Sprintf("consume:point:order:%d", orderID)

				// 1 尝试加锁 保留 24 小时防重记录
				ok, err := redis.RDB.SetNX(ctx, lockKey, 1, 24*time.Hour).Result()
				if err != nil {
					logger.Log.Error("redis error during point consumer", zap.Error(err))
					return consumer.ConsumeRetryLater, nil // Redis 出错，稍后重试
				}
				if !ok {
					// 没抢到锁,消息已经被处理过了
					logger.Log.Warn("[Point Service] duplicate message detected, safely ignored", zap.Uint("orderID", orderID))
					continue // 已经处理过这个订单的积分发放了，直接跳过
				}
				// 2 确保解锁（这里设置了过期时间，所以不需要显式删除锁）
				// 3 发放积分（这里直接用日志模拟，实际可以调用用户服务的接口）
				points := int(amount) // 1元=1积分
				logger.Log.Info("reward points granted", zap.Uint("order_id", orderID), zap.Uint("user_id", userID), zap.Int("points", points))
			}
			// 消费成功，向 MQ 返回 ACK
			return consumer.ConsumeSuccess, nil
		})
	if err != nil {
		return err
	}
	return PointConsumer.Start()
}

// ClosePointConsumer 优雅关闭消费者
func ClosePointConsumer() error {
	if PointConsumer != nil {
		err := PointConsumer.Shutdown()
		if err != nil {
			logger.Log.Error("Failed to shutdown PointConsumer", zap.Error(err))
			return err
		}
		logger.Log.Info("PointConsumer shutdown successfully")
	}
	return nil
}
