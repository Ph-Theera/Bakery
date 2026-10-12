package orders

import (
	"bakery/internal/domains/orders"
	"context"
)

type QuoteCheckoutCommand = orders.Command

type QuoteCheckoutUsecase interface {
	Handle(context.Context, QuoteCheckoutCommand) (orders.Quote, error)
}
