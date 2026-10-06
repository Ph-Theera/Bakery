package usecases

import (
	"bakery/internal/domains/product"
	"bakery/internal/ports/port_in"
	"bakery/internal/ports/port_out"
	"context"
)

type getProduct struct{ repo port_out.ProductRepository }

func NewGetProduct(repo port_out.ProductRepository) port_in.GetProductUsecase {
	return &getProduct{repo}
}
func (u *getProduct) Handle(ctx context.Context, cmd int64) (product.Data, error) {
	if cmd <= 0 {
		return product.Data{}, product.ErrInvalid
	}
	return u.repo.Get(ctx, cmd)
}
