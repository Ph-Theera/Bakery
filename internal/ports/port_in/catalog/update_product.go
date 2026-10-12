package catalog

import (
	"bakery/internal/domains/catalog"
	"context"
)

type UpdateProductCommand struct {
	ID      string
	Product catalog.Data
}

type UpdateProductUsecase interface {
	Handle(context.Context, UpdateProductCommand) (catalog.Data, error)
}
