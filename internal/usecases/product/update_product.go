package product

import (
	"bakery/internal/domains/catalog"
	portin "bakery/internal/ports/port_in/catalog"
	catalogout "bakery/internal/ports/port_out/catalog"
	inventoryout "bakery/internal/ports/port_out/inventory"
	"bakery/internal/shared"
	"context"
)

type UpdateProductArgs struct {
	ExecWithTx          shared.ExecWithTx
	CatalogRepository   catalogout.Repository
	InventoryRepository inventoryout.Repository
}
type updateProductHandler struct{ args UpdateProductArgs }

func NewUpdateProductUsecase(args UpdateProductArgs) portin.UpdateProductUsecase {
	return &updateProductHandler{args: args}
}
func (h *updateProductHandler) Handle(ctx context.Context, cmd portin.UpdateProductCommand) (catalog.Data, error) {
	if !shared.UUID(cmd.ID) {
		return catalog.Data{}, shared.ErrInvalid
	}
	p, e := catalog.Validate(cmd.Product)
	if e != nil {
		return catalog.Data{}, e
	}
	var out catalog.Data
	e = h.args.ExecWithTx.Exec(ctx, func(tx shared.Transaction) error {
		if e := h.args.CatalogRepository.Lock(ctx, tx, cmd.ID); e != nil {
			return e
		}
		reserved, e := h.args.InventoryRepository.Reserved(ctx, tx, cmd.ID)
		if e != nil {
			return e
		}
		if e = catalog.ValidateStock(int64(p.Stock), reserved); e != nil {
			return e
		}
		out, e = h.args.CatalogRepository.Save(ctx, tx, cmd.ID, p)
		return e
	})
	return out, e
}
