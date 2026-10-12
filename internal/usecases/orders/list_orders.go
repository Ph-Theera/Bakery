package orders

import (
	"bakery/internal/domains/orders"
	portin "bakery/internal/ports/port_in/orders"
	ordersout "bakery/internal/ports/port_out/orders"
	"bakery/internal/shared"
	"context"
)

type ListOrdersArgs struct{ OrderRepository ordersout.Repository }
type listOrdersHandler struct{ args ListOrdersArgs }

func NewListOrdersUsecase(args ListOrdersArgs) portin.ListOrdersUsecase {
	return &listOrdersHandler{args: args}
}
func (h *listOrdersHandler) Handle(ctx context.Context, q portin.ListOrdersQuery) ([]orders.Data, error) {
	if q.Limit < 1 || q.Limit > 100 || q.Offset < 0 {
		return nil, shared.ErrInvalid
	}
	return h.args.OrderRepository.List(ctx, q.Status, q.Limit, q.Offset)
}
