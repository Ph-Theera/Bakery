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

type SubmitPaymentArgs struct {
	ExecWithTx          shared.ExecWithTx
	OrderRepository     ordersout.Repository
	PaymentRepository   paymentsout.Repository
	InventoryRepository inventoryout.Repository
}
type submitPaymentHandler struct{ args SubmitPaymentArgs }

func NewSubmitPaymentUsecase(args SubmitPaymentArgs) portin.SubmitPaymentUsecase {
	return &submitPaymentHandler{args: args}
}
func (h *submitPaymentHandler) Handle(ctx context.Context, c portin.SubmitPaymentCommand) error {
	if !shared.UUID(c.PaymentID) || !shared.UUID(c.OrderID) {
		return shared.ErrInvalid
	}
	id, e := h.args.PaymentRepository.OrderID(ctx, c.PaymentID)
	if e != nil {
		return e
	}
	if id != c.OrderID {
		return shared.ErrForbidden
	}
	return h.args.ExecWithTx.Exec(ctx, func(tx shared.Transaction) error {
		state, e := h.args.OrderRepository.LockStatus(ctx, tx, c.OrderID)
		if e != nil {
			return e
		}
		payment, e := h.args.PaymentRepository.LockStatus(ctx, tx, c.PaymentID)
		if e != nil {
			return e
		}
		if !orders.CanSubmit(state, payment) {
			return shared.ErrConflict
		}
		yes, e := h.args.PaymentRepository.HasEvidence(ctx, tx, c.PaymentID)
		if e != nil {
			return e
		}
		if !yes {
			return shared.ErrConflict
		}
		active, e := h.args.InventoryRepository.Active(ctx, tx, c.OrderID)
		if e != nil {
			return e
		}
		if len(active) == 0 {
			return shared.ErrConflict
		}
		if e = h.args.PaymentRepository.Submit(ctx, tx, c.PaymentID); e != nil {
			return e
		}
		if e = h.args.OrderRepository.ChangeStatus(ctx, tx, c.OrderID, "pending_verification", ""); e != nil {
			return e
		}
		if state != "pending_verification" {
			return h.args.OrderRepository.History(ctx, tx, c.OrderID, state, "pending_verification", "", "")
		}
		return nil
	})
}
