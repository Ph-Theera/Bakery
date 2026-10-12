package composite

import (
	"bakery/internal/domains/orders"
	portin "bakery/internal/ports/port_in/orders"
	"context"
)

type ApprovePaymentArgs = PaymentReviewArgs
type approvePaymentHandler struct{ args ApprovePaymentArgs }

func NewApprovePaymentUsecase(args ApprovePaymentArgs) portin.ApprovePaymentUsecase {
	return &approvePaymentHandler{args: args}
}
func (h *approvePaymentHandler) Handle(ctx context.Context, c portin.ApprovePaymentCommand) (orders.Data, error) {
	return reviewPayment(ctx, h.args, c.PaymentID, "approved", c.Actor, "")
}
