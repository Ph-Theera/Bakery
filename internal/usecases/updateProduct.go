package usecases

import (
	"bakery/internal/domains/product"
	portin "bakery/internal/ports/port_in/products"
	portout "bakery/internal/ports/port_out/products"
	"context"
)

type updateProduct struct{ repo portout.ProductRepository }

func NewUpdateProduct(repo portout.ProductRepository) portin.UpdateProductUsecase {
	return &updateProduct{repo}
}
func (u *updateProduct) Handle(ctx context.Context, cmd portin.UpdateProductCommand) (product.Data, error) {
	if cmd.ID <= 0 {
		return product.Data{}, product.ErrInvalid
	}
	i, err := cmd.Input.Normalize()
	if err != nil {
		return product.Data{}, err
	}
	return u.repo.Update(ctx, cmd.ID, i)
}
