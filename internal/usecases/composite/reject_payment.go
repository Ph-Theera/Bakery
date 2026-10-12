package composite

import (
	"bakery/internal/domains/orders"
	portin "bakery/internal/ports/port_in/orders"
	"context"
)

type RejectPaymentArgs = PaymentReviewArgs
type rejectPaymentHandler struct{ args RejectPaymentArgs }

func NewRejectPaymentUsecase(args RejectPaymentArgs) portin.RejectPaymentUsecase {
	return &rejectPaymentHandler{args: args}
}
func (h *rejectPaymentHandler) Handle(ctx context.Context, c portin.RejectPaymentCommand) (orders.Data, error) {
	return reviewPayment(ctx, h.args, c.PaymentID, "rejected", c.Actor, c.Reason)
}
