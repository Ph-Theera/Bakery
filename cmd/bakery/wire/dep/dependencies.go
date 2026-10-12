package dep

import (
	"bakery/infrastructure/config"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dependencies struct{ Db *pgxpool.Pool }

func NewDependencies(ctx context.Context, c config.Config) (Dependencies, error) {
	cfg, e := pgxpool.ParseConfig(c.DatabaseURL)
	if e != nil {
		return Dependencies{}, errors.New("invalid DATABASE_URL")
	}
	cfg.MaxConns = 5
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	db, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		return Dependencies{}, errors.New("database initialization failed")
	}
	if e = db.Ping(ctx); e != nil {
		db.Close()
		return Dependencies{}, errors.New("database connection failed; check DATABASE_URL and network")
	}
	if c.AdminToken != "" {
		var allowed bool
		if e = db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM app_schema.users WHERE id::text=$1 AND role='admin' AND is_active AND deleted_at IS NULL)", c.AdminID).Scan(&allowed); e != nil || !allowed {
			db.Close()
			return Dependencies{}, errors.New("ADMIN_USER_ID must identify an active admin")
		}
	}
	return Dependencies{Db: db}, nil
}
func (d Dependencies) Close() { d.Db.Close() }
