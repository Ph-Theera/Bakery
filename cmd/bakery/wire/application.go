package wire

import (
	"bakery/infrastructure/config"
	"bakery/internal/adapters/inbound/api"
	"bakery/internal/adapters/outbound/db/catalog"
	ordersdb "bakery/internal/adapters/outbound/db/orders"
	storedb "bakery/internal/adapters/outbound/db/store"
	usersdb "bakery/internal/adapters/outbound/db/users"
	"bakery/internal/adapters/outbound/line"
	"bakery/internal/usecases/composite"
	"bakery/internal/usecases/product"
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Application(ctx context.Context, c config.Config) (http.Handler, func(), error) {
	cfg, err := pgxpool.ParseConfig(c.DatabaseURL)
	if err != nil {
		return nil, nil, errors.New("invalid DATABASE_URL")
	}
	cfg.MaxConns = 5
	// Avoid prepared statements so transaction poolers are also supported.
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, nil, errors.New("database initialization failed")
	}
	if err = db.Ping(ctx); err != nil {
		db.Close()
		return nil, nil, errors.New("database connection failed; check DATABASE_URL and network")
	}
	if c.AdminToken != "" {
		var allowed bool
		if err = db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM app_schema.users WHERE id::text=$1 AND role='admin' AND is_active AND deleted_at IS NULL)", c.AdminID).Scan(&allowed); err != nil || !allowed {
			db.Close()
			return nil, nil, errors.New("ADMIN_USER_ID must identify an active admin")
		}
	}
	handler := api.Server{Catalog: product.Service{Repo: catalog.Repository{DB: db, ImageBase: c.ImageBase}}, Orders: ordersdb.Repository{DB: db}, Store: storedb.Repository{DB: db}, Users: usersdb.Repository{DB: db}, Identity: line.Identity{Client: &http.Client{Timeout: 5 * time.Second}, ChannelID: c.LineLoginID}, Auth: api.Auth{Secret: c.AuthSecret, AdminToken: c.AdminToken, AdminID: c.AdminID}, LineSecret: c.LineSecret, AllowedOrigin: c.AllowedOrigin}.Handler()
	mux := http.NewServeMux()
	// Keep the versioned route for new clients while retaining the unversioned
	// route for existing callers and the current handler tests.
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", handler))
	mux.Handle("/", handler)
	workerCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		if c.LineToken == "" || c.LineRecipient == "" {
			return
		}
		worker := composite.Notifications{Repo: ordersdb.Repository{DB: db}, Sender: line.Sender{Client: &http.Client{Timeout: 5 * time.Second}, Token: c.LineToken, Recipient: c.LineRecipient}}
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			jobCtx, stop := context.WithTimeout(workerCtx, 15*time.Second)
			err := worker.Dispatch(jobCtx)
			stop()
			if err != nil && workerCtx.Err() == nil {
				log.Print("LINE outbox dispatch failed; messages remain queued")
			}
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return mux, func() { cancel(); <-done; db.Close() }, nil
}
