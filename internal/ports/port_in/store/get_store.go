package store

import (
	"bakery/internal/domains/store"
	"context"
)

type GetStoreQuery struct{}

type GetStoreUsecase interface {
	Handle(context.Context, GetStoreQuery) (store.Settings, error)
}
