package product

import (
	"bakery/internal/domains/catalog"
	portin "bakery/internal/ports/port_in/catalog"
	catalogout "bakery/internal/ports/port_out/catalog"
	"context"
)

type CreateProductArgs struct{ CatalogRepository catalogout.Repository }
type createProductHandler struct{ args CreateProductArgs }

func NewCreateProductUsecase(args CreateProductArgs) portin.CreateProductUsecase {
	return &createProductHandler{args: args}
}
func (h *createProductHandler) Handle(ctx context.Context, cmd portin.CreateProductCommand) (catalog.Data, error) {
	p, e := catalog.Validate(cmd.Product)
	if e != nil {
		return catalog.Data{}, e
	}
	return h.args.CatalogRepository.Save(ctx, nil, "", p)
}
