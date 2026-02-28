package dto

import "time"

type CreateOrderRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

type OrderResponse struct {
	ID        uint        `json:"id"`
	UserID    uint        `json:"user_id"`
	Amount    int64       `json:"amount"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}
type OrderStatus string
