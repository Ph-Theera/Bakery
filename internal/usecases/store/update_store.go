package store

import (
	"bakery/internal/domains/store"
	portin "bakery/internal/ports/port_in/store"
	storeout "bakery/internal/ports/port_out/store"
	"context"
)

type UpdateStoreArgs struct{ StoreRepository storeout.Repository }
type updateStoreHandler struct{ args UpdateStoreArgs }

func NewUpdateStoreUsecase(args UpdateStoreArgs) portin.UpdateStoreUsecase {
	return &updateStoreHandler{args: args}
}
func (h *updateStoreHandler) Handle(ctx context.Context, cmd portin.UpdateStoreCommand) (store.Settings, error) {
	if e := store.Validate(cmd.Settings); e != nil {
		return store.Settings{}, e
	}
	return h.args.StoreRepository.Save(ctx, cmd.Settings)
}
