package orders

import (
	"bakery/internal/domains/orders"
	"context"
)

type UpdateOrderStatusCommand struct{ ID, Status, Actor, Note string }

type UpdateOrderStatusUsecase interface {
	Handle(context.Context, UpdateOrderStatusCommand) (orders.Data, error)
}
