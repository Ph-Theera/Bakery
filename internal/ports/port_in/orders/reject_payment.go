package orders

import (
	"bakery/internal/domains/orders"
	"context"
)

type RejectPaymentCommand struct{ PaymentID, Actor, Reason string }

type RejectPaymentUsecase interface {
	Handle(context.Context, RejectPaymentCommand) (orders.Data, error)
}
