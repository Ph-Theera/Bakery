package product

import (
	"bakery/internal/domains/catalog"
	portin "bakery/internal/ports/port_in/catalog"
	catalogout "bakery/internal/ports/port_out/catalog"
	"context"
)

type ListCategoriesArgs struct{ CatalogRepository catalogout.Repository }
type listCategoriesHandler struct{ args ListCategoriesArgs }

func NewListCategoriesUsecase(args ListCategoriesArgs) portin.ListCategoriesUsecase {
	return &listCategoriesHandler{args: args}
}
func (h *listCategoriesHandler) Handle(ctx context.Context, q portin.ListCategoriesQuery) ([]catalog.Category, error) {
	return h.args.CatalogRepository.Categories(ctx)
}
