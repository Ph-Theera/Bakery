package store

import (
	"bakery/internal/domains/store"
	"bakery/internal/shared"
	"context"
)

type Repository interface {
	Get(context.Context) (store.Settings, error)
	Read(context.Context, shared.Transaction) (store.Settings, error)
	Save(context.Context, store.Settings) (store.Settings, error)
}
