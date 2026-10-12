package catalog

import (
	"bakery/internal/domains/catalog"
	"context"
)

type GetProductQuery struct {
	ID     string
	Public bool
}

type GetProductUsecase interface {
	Handle(context.Context, GetProductQuery) (catalog.Data, error)
}
