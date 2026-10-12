package ordersdb

import (
	"bakery/internal/domains/orders"
	portout "bakery/internal/ports/port_out/orders"
	"bakery/internal/shared"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ DB *pgxpool.Pool }

func NewOrderRepository(db *pgxpool.Pool) portout.Repository { return &Repository{DB: db} }

type querier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func missing(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.ErrNotFound
	}
	return err
}

const orderCols = "o.id::text,o.order_number,COALESCE(p.id::text,''),o.customer_name_snapshot,o.total_satang,o.status,COALESCE(p.status,''),o.created_at"

func get(ctx context.Context, q querier, id string) (orders.Data, error) {
	var o orders.Data
	var total int64
	err := q.QueryRow(ctx, "SELECT "+orderCols+" FROM app_schema.orders o LEFT JOIN LATERAL (SELECT id,status FROM app_schema.payments WHERE order_id=o.id AND deleted_at IS NULL ORDER BY created_at DESC,id LIMIT 1) p ON true WHERE o.id=$1 AND o.deleted_at IS NULL", id).Scan(&o.ID, &o.Number, &o.PaymentID, &o.Customer, &total, &o.Status, &o.PaymentStatus, &o.CreatedAt)
	if err != nil {
		return o, missing(err)
	}
	o.Total = float64(total) / 100
	o.Items = []orders.Line{}
	rows, err := q.Query(ctx, "SELECT product_name_snapshot,quantity,unit_price_satang FROM app_schema.order_items WHERE order_id=$1 AND deleted_at IS NULL ORDER BY id", id)
	if err != nil {
		return o, err
	}
	defer rows.Close()
	for rows.Next() {
		var l orders.Line
		var p int64
		if err = rows.Scan(&l.Name, &l.Quantity, &p); err != nil {
			return o, err
		}
		l.Price = float64(p) / 100
		o.Items = append(o.Items, l)
	}
	return o, rows.Err()
}
func (r Repository) Get(ctx context.Context, id string) (orders.Data, error) {
	return get(ctx, r.DB, id)
}
func (r Repository) List(ctx context.Context, status string, limit, offset int) ([]orders.Data, error) {
	rows, err := r.DB.Query(ctx, "SELECT id::text FROM app_schema.orders WHERE deleted_at IS NULL AND ($1='' OR status=$1) ORDER BY created_at DESC,id LIMIT $2 OFFSET $3", status, limit, offset)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	out := []orders.Data{}
	for _, id := range ids {
		o, e := get(ctx, r.DB, id)
		if e != nil {
			return nil, e
		}
		out = append(out, o)
	}
	return out, nil
}
func (r Repository) Note(ctx context.Context, id, note, actor string) (orders.Data, error) {
	tag, err := r.DB.Exec(ctx, "UPDATE app_schema.orders SET internal_note=$2,updated_by=$3 WHERE id=$1 AND deleted_at IS NULL", id, note, actor)
	if err != nil {
		return orders.Data{}, err
	}
	if tag.RowsAffected() == 0 {
		return orders.Data{}, shared.ErrNotFound
	}
	return r.Get(ctx, id)
}
func (r Repository) Dashboard(ctx context.Context) (map[string]any, error) {
	var sales int64
	var count, pending int
	err := r.DB.QueryRow(ctx, "SELECT COALESCE(sum(total_satang) FILTER (WHERE status IN ('confirmed','preparing','ready_for_pickup','shipped','completed')),0),count(*),count(*) FILTER (WHERE status='pending_verification') FROM app_schema.orders WHERE deleted_at IS NULL").Scan(&sales, &count, &pending)
	return map[string]any{"total_sales_thb": float64(sales) / 100, "total_orders": count, "pending_payments": pending}, err
}
func (r Repository) Owned(ctx context.Context, id, user string) (bool, error) {
	var yes bool
	err := r.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM app_schema.orders o JOIN app_schema.customers c ON c.id=o.customer_id WHERE o.id=$1 AND c.user_id=$2 AND o.deleted_at IS NULL AND c.deleted_at IS NULL)", id, user).Scan(&yes)
	return yes, err
}
func (r Repository) Mine(ctx context.Context, user string, limit, offset int) ([]orders.Data, error) {
	rows, err := r.DB.Query(ctx, "SELECT o.id::text FROM app_schema.orders o JOIN app_schema.customers c ON c.id=o.customer_id WHERE c.user_id=$1 AND o.deleted_at IS NULL ORDER BY o.created_at DESC,o.id LIMIT $2 OFFSET $3", user, limit, offset)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	out := []orders.Data{}
	for _, id := range ids {
		o, e := r.Get(ctx, id)
		if e != nil {
			return nil, e
		}
		out = append(out, o)
	}
	return out, nil
}
