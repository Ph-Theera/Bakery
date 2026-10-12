package composite

import (
	"bakery/internal/domains/orders"
	portin "bakery/internal/ports/port_in/orders"
	"bakery/internal/services/checkout"
	"context"
)

type QuoteCheckoutArgs struct{ Pricing checkout.Pricing }
type quoteCheckoutHandler struct{ args QuoteCheckoutArgs }

func NewQuoteCheckoutUsecase(args QuoteCheckoutArgs) portin.QuoteCheckoutUsecase {
	return &quoteCheckoutHandler{args: args}
}
func (h *quoteCheckoutHandler) Handle(ctx context.Context, c portin.QuoteCheckoutCommand) (orders.Quote, error) {
	if e := orders.ValidateQuote(c); e != nil {
		return orders.Quote{}, e
	}
	q, _, e := h.args.Pricing.Quote(ctx, nil, c, false)
	return q, e
}
