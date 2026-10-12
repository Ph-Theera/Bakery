package payments

import (
	portin "bakery/internal/ports/port_in/orders"
	paymentsout "bakery/internal/ports/port_out/payments"
	"bakery/internal/shared"
	"context"
)

type GetPaymentSlipArgs struct{ PaymentRepository paymentsout.Repository }
type getPaymentSlipHandler struct{ args GetPaymentSlipArgs }

func NewGetPaymentSlipUsecase(args GetPaymentSlipArgs) portin.GetPaymentSlipUsecase {
	return &getPaymentSlipHandler{args: args}
}
func (h *getPaymentSlipHandler) Handle(ctx context.Context, q portin.GetPaymentSlipQuery) (portin.SlipData, error) {
	if !shared.UUID(q.PaymentID) {
		return portin.SlipData{}, shared.ErrInvalid
	}
	b, m, e := h.args.PaymentRepository.Evidence(ctx, q.PaymentID)
	return portin.SlipData{Content: b, Mime: m}, e
}
