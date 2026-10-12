package composite

import (
	"bakery/internal/domains/orders"
	portin "bakery/internal/ports/port_in/orders"
	inventoryout "bakery/internal/ports/port_out/inventory"
	ordersout "bakery/internal/ports/port_out/orders"
	paymentsout "bakery/internal/ports/port_out/payments"
	"bakery/internal/shared"
	"context"
)

type UpdateOrderStatusArgs struct {
	ExecWithTx          shared.ExecWithTx
	OrderRepository     ordersout.Repository
	PaymentRepository   paymentsout.Repository
	InventoryRepository inventoryout.Repository
}
type updateOrderStatusHandler struct{ args UpdateOrderStatusArgs }

func NewUpdateOrderStatusUsecase(args UpdateOrderStatusArgs) portin.UpdateOrderStatusUsecase {
	return &updateOrderStatusHandler{args: args}
}
func (h *updateOrderStatusHandler) Handle(ctx context.Context, c portin.UpdateOrderStatusCommand) (orders.Data, error) {
	if !shared.UUID(c.ID) || !shared.UUID(c.Actor) {
		return orders.Data{}, shared.ErrInvalid
	}
	var out orders.Data
	e := h.args.ExecWithTx.Exec(ctx, func(tx shared.Transaction) error {
		from, e := h.args.OrderRepository.LockStatus(ctx, tx, c.ID)
		if e != nil {
			return e
		}
		if e = orders.ValidateTransition(from, c.Status); e != nil {
			return e
		}
		if c.Status == "cancelled" {
			if e = h.args.InventoryRepository.Release(ctx, tx, c.ID); e != nil {
				return e
			}
			if e = h.args.PaymentRepository.CancelPending(ctx, tx, c.ID); e != nil {
				return e
			}
		}
		if e = h.args.OrderRepository.ChangeStatus(ctx, tx, c.ID, c.Status, c.Actor); e != nil {
			return e
		}
		if from != c.Status {
			if e = h.args.OrderRepository.History(ctx, tx, c.ID, from, c.Status, c.Actor, c.Note); e != nil {
				return e
			}
		}
		out, e = h.args.OrderRepository.GetTx(ctx, tx, c.ID)
		return e
	})
	return out, e
}
