package orders

import (
	"bakery/internal/domains/orders"
	"bakery/internal/shared"
	"context"
)

type Repository interface {
	Get(context.Context, string) (orders.Data, error)
	List(context.Context, string, int, int) ([]orders.Data, error)
	Note(context.Context, string, string, string) (orders.Data, error)
	Dashboard(context.Context) (map[string]any, error)
	Mine(context.Context, string, int, int) ([]orders.Data, error)
	Owned(context.Context, string, string) (bool, error)
	LockKey(context.Context, shared.Transaction, string) error
	FindByKey(context.Context, shared.Transaction, string) (orders.Data, string, error)
	Insert(context.Context, shared.Transaction, orders.Command, string, orders.Quote, string) (string, error)
	InsertItems(context.Context, shared.Transaction, string, []orders.PricedItem) error
	GetTx(context.Context, shared.Transaction, string) (orders.Data, error)
	LockStatus(context.Context, shared.Transaction, string) (string, error)
	ChangeStatus(context.Context, shared.Transaction, string, string, string) error
	History(context.Context, shared.Transaction, string, string, string, string, string) error
	ItemCount(context.Context, shared.Transaction, string) (int, error)
}
