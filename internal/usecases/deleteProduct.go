package usecases

import (
	"bakery/internal/domains/product"
	portin "bakery/internal/ports/port_in/products"
	portout "bakery/internal/ports/port_out/products"
	"context"
)

type deleteProduct struct{ repo portout.ProductRepository }

func NewDeleteProduct(repo portout.ProductRepository) portin.DeleteProductUsecase {
	return &deleteProduct{repo}
}
func (u *deleteProduct) Handle(ctx context.Context, cmd int64) error {
	if cmd <= 0 {
		return product.ErrInvalid
	}
	return u.repo.Delete(ctx, cmd)
}
