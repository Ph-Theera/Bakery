package orders

import (
	"bakery/internal/domains/orders"
	"context"
)

type ListMyOrdersQuery struct {
	UserID        string
	Limit, Offset int
}

type ListMyOrdersUsecase interface {
	Handle(context.Context, ListMyOrdersQuery) ([]orders.Data, error)
}
