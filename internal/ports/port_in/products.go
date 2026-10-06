package port_in

import (
	"bakery/internal/domains/product"
	"context"
)

type CreateProductCommand = product.Input
type UpdateProductCommand struct {
	ID    int64
	Input product.Input
}
type ListProductsQuery struct {
	Limit  int
	Offset int
}
type CreateProductUsecase interface {
	Handle(context.Context, CreateProductCommand) (product.Data, error)
}
type GetProductUsecase interface {
	Handle(context.Context, int64) (product.Data, error)
}
type ListProductsUsecase interface {
	Handle(context.Context, ListProductsQuery) ([]product.Data, error)
}
type UpdateProductUsecase interface {
	Handle(context.Context, UpdateProductCommand) (product.Data, error)
}
type DeleteProductUsecase interface {
	Handle(context.Context, int64) error
}
