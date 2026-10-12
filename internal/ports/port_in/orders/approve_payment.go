package orders

import (
	"bakery/internal/domains/orders"
	"context"
)

type ApprovePaymentCommand struct{ PaymentID, Actor string }

type ApprovePaymentUsecase interface {
	Handle(context.Context, ApprovePaymentCommand) (orders.Data, error)
}
