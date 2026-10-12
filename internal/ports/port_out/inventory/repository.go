package inventory

import (
	"bakery/internal/domains/orders"
	"bakery/internal/shared"
	"context"
)

type Repository interface {
	Reserved(context.Context, shared.Transaction, string) (int64, error)
	Reserve(context.Context, shared.Transaction, string, []orders.PricedItem) error
	Active(context.Context, shared.Transaction, string) ([]orders.Item, error)
	Release(context.Context, shared.Transaction, string) error
	Commit(context.Context, shared.Transaction, string) error
	RecordSale(context.Context, shared.Transaction, string, orders.Item, string) error
}
