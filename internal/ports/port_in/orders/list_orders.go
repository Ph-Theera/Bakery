package orders

import (
	"bakery/internal/domains/orders"
	"context"
)

type ListOrdersQuery struct {
	Status        string
	Limit, Offset int
}

type ListOrdersUsecase interface {
	Handle(context.Context, ListOrdersQuery) ([]orders.Data, error)
}
