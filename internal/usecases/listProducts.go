package usecases

import (
	"bakery/internal/domains/product"
	"bakery/internal/ports/port_in"
	"bakery/internal/ports/port_out"
	"context"
)

type listProducts struct{ repo port_out.ProductRepository }

func NewListProducts(repo port_out.ProductRepository) port_in.ListProductsUsecase {
	return &listProducts{repo}
}
func (u *listProducts) Handle(ctx context.Context, cmd port_in.ListProductsQuery) ([]product.Data, error) {
	if cmd.Limit < 1 || cmd.Limit > 100 || cmd.Offset < 0 {
		return nil, product.ErrInvalid
	}
	return u.repo.List(ctx, cmd.Limit, cmd.Offset)
}
