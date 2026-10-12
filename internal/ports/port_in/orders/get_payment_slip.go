package orders

import (
	"context"
)

type GetPaymentSlipQuery struct{ PaymentID string }
type SlipData struct {
	Content []byte
	Mime    string
}
type GetPaymentSlipUsecase interface {
	Handle(context.Context, GetPaymentSlipQuery) (SlipData, error)
}
