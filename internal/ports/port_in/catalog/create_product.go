package catalog

import (
	"bakery/internal/domains/catalog"
	"context"
)

type CreateProductCommand struct{ Product catalog.Data }

type CreateProductUsecase interface {
	Handle(context.Context, CreateProductCommand) (catalog.Data, error)
}
