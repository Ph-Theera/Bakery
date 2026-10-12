package orders

import (
	"bakery/internal/domains/orders"
	"context"
)

type UpdateOrderNoteCommand struct{ ID, Note, Actor string }

type UpdateOrderNoteUsecase interface {
	Handle(context.Context, UpdateOrderNoteCommand) (orders.Data, error)
}
