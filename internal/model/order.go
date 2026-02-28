package model

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	UserID    uint        `gorm:"index"` // 与 User.ID 完全一致
	Amount    int64       // 金额（单位：分）
	Status    OrderStatus // created / paid / canceled
	RequestID string      `gorm:"size:64;uniqueIndex"`
}
type OrderStatus string

const (
	OrderStatusCreated  OrderStatus = "created"
	OrderStatusPaid     OrderStatus = "paid"
	OrderStatusCanceled OrderStatus = "canceled"
)
