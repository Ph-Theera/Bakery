package orders

import (
	"bakery/internal/domains/orders"
	portin "bakery/internal/ports/port_in/orders"
	ordersout "bakery/internal/ports/port_out/orders"
	"bakery/internal/shared"
	"context"
)

type GetOrderArgs struct{ OrderRepository ordersout.Repository }
type getOrderHandler struct{ args GetOrderArgs }

func NewGetOrderUsecase(args GetOrderArgs) portin.GetOrderUsecase {
	return &getOrderHandler{args: args}
}
func (h *getOrderHandler) Handle(ctx context.Context, q portin.GetOrderQuery) (orders.Data, error) {
	if !shared.UUID(q.ID) {
		return orders.Data{}, shared.ErrInvalid
	}
	return h.args.OrderRepository.Get(ctx, q.ID)
}
