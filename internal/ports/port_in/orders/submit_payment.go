package orders

import (
	"context"
)

type SubmitPaymentCommand struct{ PaymentID, OrderID string }

type SubmitPaymentUsecase interface {
	Handle(context.Context, SubmitPaymentCommand) error
}
