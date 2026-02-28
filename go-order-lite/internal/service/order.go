package service

import (
	"context"
	"errors"
	"fmt"
	"go-order-lite/internal/dao"
	"go-order-lite/internal/model"
	"go-order-lite/pkg/errno"
	"go-order-lite/pkg/logger"
	"go-order-lite/pkg/mysql"
	"go-order-lite/pkg/redis"
	"time"

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
	// 订单创建并落库成功后，加入 Redis 延迟队列，15分钟后触发
	err = PushOrderToDelayQueue(ctx, order.ID, 15*time.Minute)
	if err != nil {
		// 这里记录日志即可，不应该返回报错阻塞用户下单成功
		log.Error("push to delay queue failed", zap.Error(err), zap.Uint("orderID", order.ID))
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

func PayOrder(userID, orderID uint) error {
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
func SystemCancelOrder(orderID uint) error {
	order, err := dao.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	// 只有 created 状态才能被取消（防并发：可能用户刚好支付完成）
	if !canTransfer(order.Status, model.OrderStatusCanceled) {
		return nil // 状态不对不报错，直接忽略即可
	}

	order.Status = model.OrderStatusCanceled
	return dao.UpdateOrder(order)
}
