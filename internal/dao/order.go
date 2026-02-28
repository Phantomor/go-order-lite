package dao

import (
	"go-order-lite/internal/model"
	"go-order-lite/pkg/mysql"

	"gorm.io/gorm"
)

func CreateOrder(tx *gorm.DB, order *model.Order) error {
	return tx.Create(order).Error
}

func GetOrderByRequestID(requestID string) (*model.Order, error) {
	var order model.Order
	err := mysql.DB.
		Where("request_id = ?", requestID).
		First(&order).Error
	return &order, err
}
func GetOrderByID(orderID uint) (*model.Order, error) {
	var order model.Order
	err := mysql.DB.
		Where("id = ?", orderID).
		First(&order).Error
	return &order, err
}

func ListOrdersByUserID(userID uint, offset, limit int) ([]model.Order, error) {
	var orders []model.Order
	err := mysql.DB.
		Where("user_id = ?", userID).
		Order("id desc").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

func UpdateOrder(order *model.Order) error {
	return mysql.DB.Save(order).Error
}
func CancelExpiredOrders() error {
	return mysql.DB.Exec(`
		UPDATE orders
		SET status = ?
		WHERE status = ?
		  AND created_at < NOW() - INTERVAL 15 MINUTE
	`,
		model.OrderStatusCanceled,
		model.OrderStatusCreated,
	).Error
}
