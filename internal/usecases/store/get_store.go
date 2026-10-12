package store

import (
	"bakery/internal/domains/store"
	portin "bakery/internal/ports/port_in/store"
	storeout "bakery/internal/ports/port_out/store"
	"context"
)

type GetStoreArgs struct{ StoreRepository storeout.Repository }
type getStoreHandler struct{ args GetStoreArgs }

func NewGetStoreUsecase(args GetStoreArgs) portin.GetStoreUsecase {
	return &getStoreHandler{args: args}
}
func (h *getStoreHandler) Handle(ctx context.Context, q portin.GetStoreQuery) (store.Settings, error) {
	return h.args.StoreRepository.Get(ctx)
}
