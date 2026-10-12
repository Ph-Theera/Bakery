package orders

import (
	"bakery/internal/domains/orders"
	"context"
)

type UploadPaymentSlipCommand struct {
	PaymentID, OrderID, Mime string
	Content                  []byte
}

type UploadPaymentSlipUsecase interface {
	Handle(context.Context, UploadPaymentSlipCommand) (orders.Evidence, error)
}
