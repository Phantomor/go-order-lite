package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-order-lite/internal/dao"
	"go-order-lite/internal/model"
	"go-order-lite/pkg/errno"
	"go-order-lite/pkg/logger"
	"go-order-lite/pkg/mq"
	"go-order-lite/pkg/mysql"
	"go-order-lite/pkg/redis"
	"strconv"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

func CreateOrder(ctx context.Context, userID uint, amount int64, requestID string) (*model.Order, error) {
	// 日志
	log := logger.WithContext(ctx)

	log.Info("create order start",
		zap.Uint("user_id", userID),
	)

	if amount <= 0 {
		return nil, errno.InvalidAmount
	}
	lockKey := fmt.Sprintf("order:lock:%d", userID)
	// 1 尝试加锁
	ok, err := redis.RDB.SetNX(ctx, lockKey, 1, 5*time.Second).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errno.DuplicateOrder
	}
	// 2 确保解锁
	defer redis.RDB.Del(ctx, lockKey)

	// 幂等判断
	oldOrder, err := dao.GetOrderByRequestID(requestID)
	if err == nil {
		return oldOrder, nil
	}

	// 3 开启事务
	tx := mysql.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	status := model.OrderStatus("created")
	order := &model.Order{
		UserID:    userID,
		Amount:    amount,
		Status:    status,
		RequestID: requestID,
	}
	if err := dao.CreateOrder(tx, order); err != nil {
		tx.Rollback()
		return nil, err
	}
	// 以后可以在这里：
	// 扣库存
	// 扣余额

	tx.Commit()

	// RocketMQ 延迟消息逻辑
	// 1. 构造消息，topic为ORDER_DELAY_CANCEL，body为订单ID
	msg := primitive.NewMessage("ORDER_DELAY_CANCEL", []byte(strconv.Itoa(int(order.ID))))
	// 2. 设置延迟级别（比如 10 分钟后取消订单）
	msg.WithDelayTimeLevel(14)
	// 3. 发送消息
	_, err = mq.Producer.SendSync(ctx, msg)
	if err != nil {
		log.Error("failed to send delay message", zap.Error(err), zap.Uint("orderID", order.ID))
		// 发送消息失败不影响订单创建成功，所以这里不返回错误
	} else {
		log.Info("delay message sent successfully", zap.Uint("order_id", order.ID))
	}

	return order, nil
}

func ListMyOrders(userID uint, page, pageSize int) ([]model.Order, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return dao.ListOrdersByUserID(userID, offset, pageSize)
}

func PayOrder(ctx context.Context, userID, orderID uint) error {
	order, err := dao.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.UserID != userID {
		return errno.PermissionDenied
	}

	// 状态机校验
	if !canTransfer(order.Status, model.OrderStatusPaid) {
		return errors.New("order status cannot be paid")
	}

	order.Status = model.OrderStatusPaid

	// 发送支付成功MQ消息
	// 消息体
	event := map[string]interface{}{
		"order_id": order.ID,
		"user_id":  order.UserID,
		"amount":   order.Amount,
	}
	body, err := json.Marshal(event)
	msg := primitive.NewMessage("ORDER_PAID", body)

	//同步发送消息
	_, err = mq.Producer.SendSync(ctx, msg)
	if err != nil {
		logger.Log.Error("failed to send order paid message", zap.Error(err), zap.Uint("order_id", order.ID))
		// 发送消息失败不影响订单支付成功，所以这里不返回错误
	} else {
		logger.Log.Info("order paid message sent successfully", zap.Uint("order_id", order.ID))
	}

	return dao.UpdateOrder(order)
}
func CancelOrder(userID, orderID uint) error {
	order, err := dao.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.UserID != userID {
		return errno.PermissionDenied
	}

	if !canTransfer(order.Status, model.OrderStatusCanceled) {
		return errors.New("order status cannot be canceled")
	}

	order.Status = model.OrderStatusCanceled
	return dao.UpdateOrder(order)
}

// SystemCancelOrder 系统自动取消超时订单（无需校验 UserID）
func SystemCancelOrder(orderID uint) (bool, error) {
	order, err := dao.GetOrderByID(orderID)
	if err != nil {
		return false, err
	}

	// 只有 created 状态才能被取消（防并发：可能用户刚好支付完成）
	if !canTransfer(order.Status, model.OrderStatusCanceled) {
		return false, nil // 状态不对不报错，直接忽略即可
	}

	order.Status = model.OrderStatusCanceled
	return true, dao.UpdateOrder(order)
}
