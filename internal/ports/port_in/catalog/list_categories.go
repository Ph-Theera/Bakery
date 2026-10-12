package catalog

import (
	"bakery/internal/domains/catalog"
	"context"
)

type ListCategoriesQuery struct{}

type ListCategoriesUsecase interface {
	Handle(context.Context, ListCategoriesQuery) ([]catalog.Category, error)
}
