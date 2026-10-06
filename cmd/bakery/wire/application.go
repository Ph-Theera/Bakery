package wire

import (
	"bakery/infrastructure/config"
	"bakery/internal/adapters/inbound/rest"
	productdb "bakery/internal/adapters/outbound/db/product"
	"bakery/internal/usecases"
	"context"
	"errors"
	"net/http"

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
	repo := productdb.New(db)
	return rest.New(usecases.NewCreateProduct(repo), usecases.NewGetProduct(repo), usecases.NewListProducts(repo), usecases.NewUpdateProduct(repo), usecases.NewDeleteProduct(repo)), db.Close, nil
}
