package usecases

import (
	"bakery/internal/domains/product"
	"bakery/internal/ports/port_in"
	"bakery/internal/ports/port_out"
	"context"
)

type deleteProduct struct{ repo port_out.ProductRepository }

func NewDeleteProduct(repo port_out.ProductRepository) port_in.DeleteProductUsecase {
	return &deleteProduct{repo}
}
func (u *deleteProduct) Handle(ctx context.Context, cmd int64) error {
	if cmd <= 0 {
		return product.ErrInvalid
	}
	return u.repo.Delete(ctx, cmd)
}
