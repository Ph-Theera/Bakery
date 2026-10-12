package orders

import (
	"bakery/internal/domains/orders"
	portin "bakery/internal/ports/port_in/orders"
	ordersout "bakery/internal/ports/port_out/orders"
	"bakery/internal/shared"
	"context"
)

type ListMyOrdersArgs struct{ OrderRepository ordersout.Repository }
type listMyOrdersHandler struct{ args ListMyOrdersArgs }

func NewListMyOrdersUsecase(args ListMyOrdersArgs) portin.ListMyOrdersUsecase {
	return &listMyOrdersHandler{args: args}
}
func (h *listMyOrdersHandler) Handle(ctx context.Context, q portin.ListMyOrdersQuery) ([]orders.Data, error) {
	if !shared.UUID(q.UserID) || q.Limit < 1 || q.Limit > 100 || q.Offset < 0 {
		return nil, shared.ErrInvalid
	}
	return h.args.OrderRepository.Mine(ctx, q.UserID, q.Limit, q.Offset)
}
