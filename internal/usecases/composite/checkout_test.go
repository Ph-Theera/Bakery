package composite

import (
	"bakery/internal/shared"
	"bakery/internal/usecases/orders"
	"context"
	"errors"
	"testing"
)

type checkoutRepo struct {
	orders.Repository
	calls int
	err   error
}

func (r *checkoutRepo) Quote(context.Context, orders.Command) (orders.Quote, error) {
	r.calls++
	return orders.Quote{}, r.err
}
func TestQuoteValidationAndErrors(t *testing.T) {
	var c orders.Command
	c.Fulfillment.Type = "delivery"
	c.Fulfillment.Address = "Bangkok"
	c.Items = []orders.Item{{ProductID: "30000000-0000-4000-8000-000000000001", Quantity: 1}}
	r := &checkoutRepo{err: errors.New("database unavailable")}
	s := Checkout{Repo: r}
	if _, err := s.Quote(context.Background(), c); !errors.Is(err, r.err) || r.calls != 1 {
		t.Fatal(err)
	}
	c.Items = append(c.Items, c.Items[0])
	if _, err := s.Quote(context.Background(), c); !errors.Is(err, shared.ErrInvalid) || r.calls != 1 {
		t.Fatal("duplicate product reached repository", err)
	}
}
