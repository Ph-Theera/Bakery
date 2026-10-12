package orders

import (
	"bakery/internal/domains/orders"
	"context"
)

type CreateOrderCommand = orders.Command

type CreateOrderUsecase interface {
	Handle(context.Context, CreateOrderCommand) (orders.Data, error)
}
