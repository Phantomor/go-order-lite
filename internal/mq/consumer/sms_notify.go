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

var SMSConsumer rocketmq.PushConsumer

// InitSMSConsumer 初始化并启动短信通知的消费者
func InitSMSConsumer() error {
	var err error
	SMSConsumer, err = rocketmq.NewPushConsumer(
		consumer.WithNameServer(config.Cfg.RocketMQ.NameServers),
		consumer.WithMaxReconsumeTimes(3),
		consumer.WithGroupName("sms-notify-group"),
	)
	if err != nil {
		return err
	}
	err = SMSConsumer.Subscribe("ORDER_PAID", consumer.MessageSelector{},
		func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			// 模拟发送短信通知
			for _, msg := range msgs {
				var event map[string]interface{}
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					logger.Log.Error("invalid message body", zap.String("body", string(msg.Body)), zap.Error(err))
					continue // 数据格式错误，没必要重试，直接跳过
				}
				// 模拟业务逻辑：发送短信通知用户订单已支付
				orderID := uint(event["order_id"].(float64))

				// 幂等性校验
				lockKey := fmt.Sprintf("consume:sms:order:%d", orderID)

				// 1 尝试加锁 保留 24 小时防重记录
				ok, err := redis.RDB.SetNX(ctx, lockKey, 1, 24*time.Hour).Result()
				if err != nil {
					logger.Log.Error("redis error during SMS consumer", zap.Error(err))
					return consumer.ConsumeRetryLater, nil // Redis 出错，稍后重试
				}
				if !ok {
					logger.Log.Warn("[SMS Service] duplicate message detected, safely ignored", zap.Uint("orderID", orderID))
					continue
				}
				// 2 确保解锁（这里设置了过期时间，所以不需要显式删除锁）
				// 3 发送短信（这里直接用日志模拟，实际可以调用第三方短信服务的 SDK 来发送短信）

				userID := uint(event["user_id"].(float64))
				logger.Log.Info("sending SMS notification", zap.Uint("user_id", userID), zap.Uint("order_id", orderID))
				// 这里可以调用第三方短信服务的 SDK 来发送短信，暂时用日志模拟
				logger.Log.Info("SMS notification sent", zap.String("message_body", string(msg.Body)))
			}
			return consumer.ConsumeSuccess, nil
		},
	)
	if err != nil {
		return err
	}
	return SMSConsumer.Start()
}

// CloseSMSConsumer 优雅关闭消费者
func CloseSMSConsumer() {
	if SMSConsumer != nil {
		SMSConsumer.Shutdown()
	}
}
