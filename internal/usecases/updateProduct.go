package usecases

import (
	"bakery/internal/domains/product"
	"bakery/internal/ports/port_in"
	"bakery/internal/ports/port_out"
	"context"
)

type updateProduct struct{ repo port_out.ProductRepository }

func NewUpdateProduct(repo port_out.ProductRepository) port_in.UpdateProductUsecase {
	return &updateProduct{repo}
}
func (u *updateProduct) Handle(ctx context.Context, cmd port_in.UpdateProductCommand) (product.Data, error) {
	if cmd.ID <= 0 {
		return product.Data{}, product.ErrInvalid
	}
	i, err := cmd.Input.Normalize()
	if err != nil {
		return product.Data{}, err
	}
	return u.repo.Update(ctx, cmd.ID, i)
}
