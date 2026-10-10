package usecases

import (
	"bakery/internal/domains/product"
	portin "bakery/internal/ports/port_in/products"
	portout "bakery/internal/ports/port_out/products"
	"context"
)

type getProduct struct{ repo portout.ProductRepository }

func NewGetProduct(repo portout.ProductRepository) portin.GetProductUsecase {
	return &getProduct{repo}
}
func (u *getProduct) Handle(ctx context.Context, cmd int64) (product.Data, error) {
	if cmd <= 0 {
		return product.Data{}, product.ErrInvalid
	}
	return u.repo.Get(ctx, cmd)
}
