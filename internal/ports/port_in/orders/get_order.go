package orders

import (
	"bakery/internal/domains/orders"
	"context"
)

type GetOrderQuery struct{ ID string }

type GetOrderUsecase interface {
	Handle(context.Context, GetOrderQuery) (orders.Data, error)
}
