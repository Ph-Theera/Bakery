package product

import (
	portin "bakery/internal/ports/port_in/catalog"
	catalogout "bakery/internal/ports/port_out/catalog"
	"bakery/internal/shared"
	"context"
)

type DeleteProductArgs struct{ CatalogRepository catalogout.Repository }
type deleteProductHandler struct{ args DeleteProductArgs }

func NewDeleteProductUsecase(args DeleteProductArgs) portin.DeleteProductUsecase {
	return &deleteProductHandler{args: args}
}
func (h *deleteProductHandler) Handle(ctx context.Context, cmd portin.DeleteProductCommand) error {
	if !shared.UUID(cmd.ID) || !shared.UUID(cmd.Actor) {
		return shared.ErrInvalid
	}
	return h.args.CatalogRepository.Delete(ctx, cmd.ID, cmd.Actor)
}
