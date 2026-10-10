package products

import (
	"bakery/internal/domains/product"
	"context"
)

type ProductRepository interface {
	Create(context.Context, product.Input) (product.Data, error)
	Get(context.Context, int64) (product.Data, error)
	List(context.Context, int, int) ([]product.Data, error)
	Update(context.Context, int64, product.Input) (product.Data, error)
	Delete(context.Context, int64) error
}
