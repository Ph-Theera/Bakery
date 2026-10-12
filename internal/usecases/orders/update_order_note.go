package orders

import (
	"bakery/internal/domains/orders"
	portin "bakery/internal/ports/port_in/orders"
	ordersout "bakery/internal/ports/port_out/orders"
	"bakery/internal/shared"
	"context"
)

type UpdateOrderNoteArgs struct{ OrderRepository ordersout.Repository }
type updateOrderNoteHandler struct{ args UpdateOrderNoteArgs }

func NewUpdateOrderNoteUsecase(args UpdateOrderNoteArgs) portin.UpdateOrderNoteUsecase {
	return &updateOrderNoteHandler{args: args}
}
func (h *updateOrderNoteHandler) Handle(ctx context.Context, cmd portin.UpdateOrderNoteCommand) (orders.Data, error) {
	if !shared.UUID(cmd.ID) || !shared.UUID(cmd.Actor) || len(cmd.Note) > 5000 {
		return orders.Data{}, shared.ErrInvalid
	}
	return h.args.OrderRepository.Note(ctx, cmd.ID, cmd.Note, cmd.Actor)
}
