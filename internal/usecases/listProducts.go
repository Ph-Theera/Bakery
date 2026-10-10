package usecases

import (
	"bakery/internal/domains/product"
	portin "bakery/internal/ports/port_in/products"
	portout "bakery/internal/ports/port_out/products"
	"context"
)

type listProducts struct{ repo portout.ProductRepository }

func NewListProducts(repo portout.ProductRepository) portin.ListProductsUsecase {
	return &listProducts{repo}
}
func (u *listProducts) Handle(ctx context.Context, cmd portin.ListProductsQuery) ([]product.Data, error) {
	if cmd.Limit < 1 || cmd.Limit > 100 || cmd.Offset < 0 {
		return nil, product.ErrInvalid
	}
	return u.repo.List(ctx, cmd.Limit, cmd.Offset)
}
