package payments

import (
	portin "bakery/internal/ports/port_in/orders"
	paymentsout "bakery/internal/ports/port_out/payments"
	"bakery/internal/shared"
	"context"
)

type GetPaymentOrderArgs struct{ PaymentRepository paymentsout.Repository }
type getPaymentOrderHandler struct{ args GetPaymentOrderArgs }

func NewGetPaymentOrderUsecase(args GetPaymentOrderArgs) portin.GetPaymentOrderUsecase {
	return &getPaymentOrderHandler{args: args}
}
func (h *getPaymentOrderHandler) Handle(ctx context.Context, q portin.GetPaymentOrderQuery) (string, error) {
	if !shared.UUID(q.PaymentID) {
		return "", shared.ErrInvalid
	}
	return h.args.PaymentRepository.OrderID(ctx, q.PaymentID)
}
