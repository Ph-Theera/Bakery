package transaction

import (
	"bakery/internal/shared"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type scope struct{ tx pgx.Tx }

func (*scope) TransactionScope() {}

type Queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func Executor(db *pgxpool.Pool, tx shared.Transaction) Queryer {
	if tx == nil {
		return db
	}
	return tx.(*scope).tx
}

type executor struct{ db *pgxpool.Pool }

func NewExecWithTx(db *pgxpool.Pool) shared.ExecWithTx { return &executor{db: db} }
func (e *executor) Exec(ctx context.Context, fn func(shared.Transaction) error) error {
	tx, err := e.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if err = fn(&scope{tx: tx}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
