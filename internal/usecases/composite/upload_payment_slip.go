package composite

import (
	"bakery/internal/domains/orders"
	portin "bakery/internal/ports/port_in/orders"
	ordersout "bakery/internal/ports/port_out/orders"
	paymentsout "bakery/internal/ports/port_out/payments"
	"bakery/internal/shared"
	"context"
)

type UploadPaymentSlipArgs struct {
	ExecWithTx        shared.ExecWithTx
	OrderRepository   ordersout.Repository
	PaymentRepository paymentsout.Repository
}
type uploadPaymentSlipHandler struct{ args UploadPaymentSlipArgs }

func NewUploadPaymentSlipUsecase(args UploadPaymentSlipArgs) portin.UploadPaymentSlipUsecase {
	return &uploadPaymentSlipHandler{args: args}
}
func (h *uploadPaymentSlipHandler) Handle(ctx context.Context, c portin.UploadPaymentSlipCommand) (orders.Evidence, error) {
	if !shared.UUID(c.PaymentID) || !shared.UUID(c.OrderID) || len(c.Content) == 0 || len(c.Content) > 5<<20 {
		return orders.Evidence{}, shared.ErrInvalid
	}
	mime := c.Mime
	if mime != "image/jpeg" && mime != "image/png" && mime != "image/webp" {
		return orders.Evidence{}, shared.ErrInvalid
	}
	id, e := h.args.PaymentRepository.OrderID(ctx, c.PaymentID)
	if e != nil {
		return orders.Evidence{}, e
	}
	if id != c.OrderID {
		return orders.Evidence{}, shared.ErrForbidden
	}
	var out orders.Evidence
	e = h.args.ExecWithTx.Exec(ctx, func(tx shared.Transaction) error {
		state, e := h.args.OrderRepository.LockStatus(ctx, tx, c.OrderID)
		if e != nil {
			return e
		}
		payment, e := h.args.PaymentRepository.LockStatus(ctx, tx, c.PaymentID)
		if e != nil {
			return e
		}
		if !orders.CanSubmit(state, payment) || payment == "submitted" {
			return shared.ErrConflict
		}
		out, e = h.args.PaymentRepository.Slip(ctx, tx, c.PaymentID, mime, c.Content)
		return e
	})
	return out, e
}
