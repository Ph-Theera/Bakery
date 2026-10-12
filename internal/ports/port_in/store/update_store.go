package store

import (
	"bakery/internal/domains/store"
	"context"
)

type UpdateStoreCommand struct{ Settings store.Settings }

type UpdateStoreUsecase interface {
	Handle(context.Context, UpdateStoreCommand) (store.Settings, error)
}
