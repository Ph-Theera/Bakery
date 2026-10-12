package catalog

import (
	"bakery/internal/adapters/outbound/db/transaction"
	product "bakery/internal/domains/catalog"
	"bakery/internal/shared"
	"context"
)

func (r Repository) Lock(ctx context.Context, tx shared.Transaction, id string) error {
	var result string
	e := transaction.Executor(r.DB, tx).QueryRow(ctx, "SELECT id::text FROM app_schema.products WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", id).Scan(&result)
	return dbError(e)
}
func (r Repository) Price(ctx context.Context, tx shared.Transaction, id string, lock bool) (product.Stock, error) {
	var p product.Stock
	p.ID = id
	sql := "SELECT name,price_satang,stock_quantity FROM app_schema.products WHERE id=$1 AND is_active AND deleted_at IS NULL"
	if lock {
		sql += " FOR UPDATE"
	}
	e := transaction.Executor(r.DB, tx).QueryRow(ctx, sql, id).Scan(&p.Name, &p.Price, &p.Quantity)
	return p, dbError(e)
}
func (r Repository) Deduct(ctx context.Context, tx shared.Transaction, id string, quantity int) error {
	tag, e := transaction.Executor(r.DB, tx).Exec(ctx, "UPDATE app_schema.products SET stock_quantity=stock_quantity-$2 WHERE id=$1 AND stock_quantity>=$2", id, quantity)
	if e != nil {
		return dbError(e)
	}
	if tag.RowsAffected() != 1 {
		return shared.ErrConflict
	}
	return nil
}
