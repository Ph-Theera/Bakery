package catalog

import (
	"bakery/internal/domains/catalog"
	"context"
)

type ListProductsQuery = catalog.Query

type ListProductsUsecase interface {
	Handle(context.Context, ListProductsQuery) ([]catalog.Data, error)
}
