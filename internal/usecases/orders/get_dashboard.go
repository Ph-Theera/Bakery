package orders

import (
	portin "bakery/internal/ports/port_in/orders"
	ordersout "bakery/internal/ports/port_out/orders"
	"context"
)

type GetDashboardArgs struct{ OrderRepository ordersout.Repository }
type getDashboardHandler struct{ args GetDashboardArgs }

func NewGetDashboardUsecase(args GetDashboardArgs) portin.GetDashboardUsecase {
	return &getDashboardHandler{args: args}
}
func (h *getDashboardHandler) Handle(ctx context.Context, q portin.GetDashboardQuery) (map[string]any, error) {
	return h.args.OrderRepository.Dashboard(ctx)
}
