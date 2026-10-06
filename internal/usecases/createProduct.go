package usecases

import (
	"bakery/internal/domains/product"
	"bakery/internal/ports/port_in"
	"bakery/internal/ports/port_out"
	"context"
)

type createProduct struct{ repo port_out.ProductRepository }

func NewCreateProduct(repo port_out.ProductRepository) port_in.CreateProductUsecase {
	return &createProduct{repo}
}
func (u *createProduct) Handle(ctx context.Context, cmd port_in.CreateProductCommand) (product.Data, error) {
	i, err := cmd.Normalize()
	if err != nil {
		return product.Data{}, err
	}
	return u.repo.Create(ctx, i)
}
