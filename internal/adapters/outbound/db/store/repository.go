package storedb

import (
	"bakery/internal/adapters/outbound/db/transaction"
	"bakery/internal/domains/store"
	portout "bakery/internal/ports/port_out/store"
	"bakery/internal/shared"
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ DB *pgxpool.Pool }

func (r Repository) Get(ctx context.Context) (store.Settings, error) { return r.Read(ctx, nil) }
func (r Repository) Read(ctx context.Context, tx shared.Transaction) (store.Settings, error) {
	var raw []byte
	err := transaction.Executor(r.DB, tx).QueryRow(ctx, "SELECT value FROM app_schema.store_settings WHERE key='store'").Scan(&raw)
	var s store.Settings
	if err == nil {
		err = json.Unmarshal(raw, &s)
	}
	return s, err
}
func (r Repository) Save(ctx context.Context, s store.Settings) (store.Settings, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return s, err
	}
	_, err = r.DB.Exec(ctx, "UPDATE app_schema.store_settings SET value=$1::jsonb,updated_at=now() WHERE key='store'", string(raw))
	return s, err
}
func NewStoreRepository(db *pgxpool.Pool) portout.Repository { return &Repository{DB: db} }
