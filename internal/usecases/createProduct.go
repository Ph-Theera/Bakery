package usecases

import (
	"bakery/internal/domains/product"
	portin "bakery/internal/ports/port_in/products"
	portout "bakery/internal/ports/port_out/products"
	"context"
)

type createProduct struct{ repo portout.ProductRepository }

func NewCreateProduct(repo portout.ProductRepository) portin.CreateProductUsecase {
	return &createProduct{repo}
}
func (u *createProduct) Handle(ctx context.Context, cmd portin.CreateProductCommand) (product.Data, error) {
	i, err := cmd.Normalize()
	if err != nil {
		return product.Data{}, err
	}
	return u.repo.Create(ctx, i)
}
