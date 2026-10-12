package product

import (
	"bakery/internal/domains/catalog"
	portin "bakery/internal/ports/port_in/catalog"
	catalogout "bakery/internal/ports/port_out/catalog"
	"bakery/internal/shared"
	"context"
)

type ListProductsArgs struct{ CatalogRepository catalogout.Repository }
type listProductsHandler struct{ args ListProductsArgs }

func NewListProductsUsecase(args ListProductsArgs) portin.ListProductsUsecase {
	return &listProductsHandler{args: args}
}
func (h *listProductsHandler) Handle(ctx context.Context, q portin.ListProductsQuery) ([]catalog.Data, error) {
	if q.Limit < 1 || q.Limit > 100 || q.Offset < 0 || q.Category != "" && !shared.UUID(q.Category) {
		return nil, shared.ErrInvalid
	}
	return h.args.CatalogRepository.List(ctx, q)
}
