package product

import (
	"bakery/internal/domains/catalog"
	portin "bakery/internal/ports/port_in/catalog"
	catalogout "bakery/internal/ports/port_out/catalog"
	"bakery/internal/shared"
	"context"
)

type GetProductArgs struct{ CatalogRepository catalogout.Repository }
type getProductHandler struct{ args GetProductArgs }

func NewGetProductUsecase(args GetProductArgs) portin.GetProductUsecase {
	return &getProductHandler{args: args}
}
func (h *getProductHandler) Handle(ctx context.Context, q portin.GetProductQuery) (catalog.Data, error) {
	if q.ID == "" || len(q.ID) > 220 {
		return catalog.Data{}, shared.ErrInvalid
	}
	return h.args.CatalogRepository.Get(ctx, q.ID, q.Public)
}
