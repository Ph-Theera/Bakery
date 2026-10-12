package orders

import (
	portin "bakery/internal/ports/port_in/orders"
	ordersout "bakery/internal/ports/port_out/orders"
	"bakery/internal/shared"
	"context"
)

type CheckOrderAccessArgs struct{ OrderRepository ordersout.Repository }
type checkOrderAccessHandler struct{ args CheckOrderAccessArgs }

func NewCheckOrderAccessUsecase(args CheckOrderAccessArgs) portin.CheckOrderAccessUsecase {
	return &checkOrderAccessHandler{args: args}
}
func (h *checkOrderAccessHandler) Handle(ctx context.Context, q portin.CheckOrderAccessQuery) (bool, error) {
	if !shared.UUID(q.ID) || !shared.UUID(q.UserID) {
		return false, shared.ErrForbidden
	}
	return h.args.OrderRepository.Owned(ctx, q.ID, q.UserID)
}
