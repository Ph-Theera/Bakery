package orders

import (
	"context"
)

type GetPaymentOrderQuery struct{ PaymentID string }

type GetPaymentOrderUsecase interface {
	Handle(context.Context, GetPaymentOrderQuery) (string, error)
}
