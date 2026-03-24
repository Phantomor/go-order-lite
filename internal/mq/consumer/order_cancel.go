package consumer

import (
	"context"
	"go-order-lite/internal/service"
	"go-order-lite/pkg/config"
	"go-order-lite/pkg/logger"
	"strconv"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

var OrderCancelConsumer rocketmq.PushConsumer

// InitOrderCancelConsumer 初始化并启动延迟取消订单的消费者
func InitOrderCancelConsumer() error {
	var err error
	OrderCancelConsumer, err = rocketmq.NewPushConsumer(
		consumer.WithNameServer(config.Cfg.RocketMQ.NameServers),
		consumer.WithRetry(config.Cfg.RocketMQ.Retry),
		consumer.WithGroupName("order-cancel-group"),
	)
	if err != nil {
		return err
	}
	// 订阅延迟取消的主题
	err = OrderCancelConsumer.Subscribe("ORDER_DELAY_CANCEL", consumer.MessageSelector{},
		func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, msg := range msgs {
				orderID, err := strconv.Atoi(string(msg.Body))
				if err != nil {
					logger.Log.Error("invalid order ID in message", zap.String("body", string(msg.Body)), zap.Error(err))
					continue // 数据格式错误，没必要重试，直接跳过
				}
				// 复用 V1.0 中写好的取消订单核心逻辑
				isCanceled, err := service.SystemCancelOrder(uint(orderID))
				if err != nil {
					logger.Log.Error("failed to cancel order", zap.Int("order_id", orderID), zap.Error(err))
					return consumer.ConsumeRetryLater, nil // 处理失败，稍后重试
				}
				if isCanceled {
					// 真正把 created 改成了 canceled
					logger.Log.Info("order auto canceled successfully via rocketmq", zap.Uint("orderID", uint(orderID)))
				} else {
					// 订单已经是 paid 或者 canceled，被状态机拦截
					logger.Log.Info("delay cancel ignored: order already paid or processed", zap.Uint("orderID", uint(orderID)))
				}
			}
			// 消费成功，向 MQ 返回 ACK
			return consumer.ConsumeSuccess, nil
		})
	if err != nil {
		return err
	}
	return OrderCancelConsumer.Start()
}

// CloseOrderCancelConsumer 优雅关闭消费者
func CloseOrderCancelConsumer() error {
	if OrderCancelConsumer != nil {
		err := OrderCancelConsumer.Shutdown()
		if err != nil {
			logger.Log.Error("Failed to shutdown OrderCancelConsumer", zap.Error(err))
			return err
		}
		logger.Log.Info("OrderCancelConsumer shutdown successfully")
	}
	return nil
}
