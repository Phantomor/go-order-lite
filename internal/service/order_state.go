package service

import "go-order-lite/internal/model"

var orderStateMachine = map[model.OrderStatus][]model.OrderStatus{
	model.OrderStatusCreated: {
		model.OrderStatusPaid,
		model.OrderStatusCanceled,
	},
	model.OrderStatusPaid: {
		// 后续可以加 refund
	},
	model.OrderStatusCanceled: {
		// 终态，不允许任何转移
	},
}

func canTransfer(from, to model.OrderStatus) bool {
	nextStates, ok := orderStateMachine[from]
	if !ok {
		return false
	}
	for _, s := range nextStates {
		if s == to {
			return true
		}
	}
	return false
}
