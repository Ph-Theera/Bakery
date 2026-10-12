package catalog

import (
	"bakery/internal/domains/catalog"
	"bakery/internal/shared"
	"context"
)

type Repository interface {
	List(context.Context, catalog.Query) ([]catalog.Data, error)
	Get(context.Context, string, bool) (catalog.Data, error)
	Save(context.Context, shared.Transaction, string, catalog.Data) (catalog.Data, error)
	Delete(context.Context, string, string) error
	Categories(context.Context) ([]catalog.Category, error)
	Lock(context.Context, shared.Transaction, string) error
	Price(context.Context, shared.Transaction, string, bool) (catalog.Stock, error)
	Deduct(context.Context, shared.Transaction, string, int) error
}
