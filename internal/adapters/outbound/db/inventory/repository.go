package inventorydb

import (
	"bakery/internal/adapters/outbound/db/transaction"
	"bakery/internal/domains/orders"
	portout "bakery/internal/ports/port_out/inventory"
	"bakery/internal/shared"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ DB *pgxpool.Pool }

func NewInventoryRepository(db *pgxpool.Pool) portout.Repository { return &Repository{DB: db} }
func (r Repository) Reserved(ctx context.Context, tx shared.Transaction, id string) (int64, error) {
	var n int64
	e := transaction.Executor(r.DB, tx).QueryRow(ctx, "SELECT COALESCE(sum(quantity),0) FROM app_schema.inventory_reservations WHERE product_id=$1 AND status='reserved' AND expires_at>now() AND deleted_at IS NULL", id).Scan(&n)
	return n, e
}
func (r Repository) Reserve(ctx context.Context, tx shared.Transaction, id string, items []orders.PricedItem) error {
	for _, p := range items {
		_, e := transaction.Executor(r.DB, tx).Exec(ctx, "INSERT INTO app_schema.inventory_reservations(order_id,product_id,quantity,expires_at) VALUES($1,$2,$3,now()+interval '30 minutes')", id, p.ID, p.Quantity)
		if e != nil {
			return e
		}
	}
	return nil
}
func (r Repository) Active(ctx context.Context, tx shared.Transaction, id string) ([]orders.Item, error) {
	rows, e := transaction.Executor(r.DB, tx).Query(ctx, "SELECT product_id::text,quantity FROM app_schema.inventory_reservations WHERE order_id=$1 AND status='reserved' AND expires_at>now() AND deleted_at IS NULL ORDER BY product_id", id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []orders.Item{}
	for rows.Next() {
		var i orders.Item
		if e = rows.Scan(&i.ProductID, &i.Quantity); e != nil {
			return nil, e
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
func (r Repository) Release(ctx context.Context, tx shared.Transaction, id string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "UPDATE app_schema.inventory_reservations SET status='released',released_at=now() WHERE order_id=$1 AND status='reserved'", id)
	return e
}
func (r Repository) Commit(ctx context.Context, tx shared.Transaction, id string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "UPDATE app_schema.inventory_reservations SET status='committed' WHERE order_id=$1 AND status='reserved'", id)
	return e
}
func (r Repository) RecordSale(ctx context.Context, tx shared.Transaction, id string, i orders.Item, actor string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "INSERT INTO app_schema.inventory_movements(product_id,quantity_delta,reason,reference_type,reference_id,idempotency_key,created_by) VALUES($1,$2,'sale','order',$3,$4,$5)", i.ProductID, -i.Quantity, id, id+":"+i.ProductID, actor)
	return e
}
