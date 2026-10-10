package composite

import (
	"bakery/internal/shared"
	"bakery/internal/usecases/orders"
	"context"
	"strings"
)

type Checkout struct{ Repo orders.Repository }

func validate(c orders.Command) error {
	if len(c.Items) == 0 || len(c.Items) > 100 || (c.Fulfillment.Type != "delivery" && c.Fulfillment.Type != "pickup") || c.Fulfillment.Type == "delivery" && (strings.TrimSpace(c.Fulfillment.Address) == "" || len(c.Fulfillment.Address) > 2000) || c.Fulfillment.Type == "pickup" && c.Fulfillment.PickupAt == nil {
		return shared.ErrInvalid
	}
	seen := map[string]bool{}
	for _, i := range c.Items {
		if !shared.UUID(i.ProductID) || i.Quantity < 1 || i.Quantity > 1000 || seen[i.ProductID] {
			return shared.ErrInvalid
		}
		seen[i.ProductID] = true
	}
	return nil
}
func (s Checkout) Quote(ctx context.Context, c orders.Command) (orders.Quote, error) {
	if err := validate(c); err != nil {
		return orders.Quote{}, err
	}
	return s.Repo.Quote(ctx, c)
}
func (s Checkout) Create(ctx context.Context, c orders.Command) (orders.Data, error) {
	if err := validate(c); err != nil {
		return orders.Data{}, err
	}
	if strings.TrimSpace(c.Customer.Name) == "" || len([]rune(c.Customer.Name)) > 160 || strings.TrimSpace(c.Customer.Phone) == "" || len(c.Customer.Phone) > 40 || len(c.Key) < 16 || len(c.Key) > 100 || (c.PaymentMethod != "bank_transfer" && c.PaymentMethod != "promptpay") {
		return orders.Data{}, shared.ErrInvalid
	}
	return s.Repo.Create(ctx, c)
}
