package types

import (
	"context"

	"github.com/sangkips/order-processing-system/services/common/genproto/orders/orders"
)

type OrderService interface {
	CreateOrder(context.Context, *orders.Order) error
	GetOrders(context.Context) []*orders.Order
}
