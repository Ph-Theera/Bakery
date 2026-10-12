package ordersdb

import (
	"bakery/internal/adapters/outbound/db/transaction"
	"bakery/internal/domains/orders"
	"bakery/internal/shared"
	"context"
	"encoding/json"
	"math"
)

func (r Repository) LockKey(ctx context.Context, tx shared.Transaction, key string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", key)
	return e
}
func (r Repository) FindByKey(ctx context.Context, tx shared.Transaction, key string) (orders.Data, string, error) {
	var id, hash string
	q := transaction.Executor(r.DB, tx)
	e := q.QueryRow(ctx, "SELECT id::text,COALESCE(request_hash,'') FROM app_schema.orders WHERE idempotency_key=$1", key).Scan(&id, &hash)
	if e != nil {
		return orders.Data{}, hash, missing(e)
	}
	o, e := get(ctx, q, id)
	return o, hash, e
}
func (r Repository) Insert(ctx context.Context, tx shared.Transaction, c orders.Command, customer string, quote orders.Quote, hash string) (string, error) {
	var address any
	if c.Fulfillment.Type == "delivery" {
		b, e := json.Marshal(map[string]string{"address": c.Fulfillment.Address})
		if e != nil {
			return "", e
		}
		address = string(b)
	}
	var id string
	e := transaction.Executor(r.DB, tx).QueryRow(ctx, "INSERT INTO app_schema.orders(order_number,idempotency_key,customer_id,customer_name_snapshot,customer_phone_snapshot,fulfillment_type,delivery_address_snapshot,pickup_at,subtotal_satang,delivery_fee_satang,total_satang,placed_at,request_hash) VALUES('TB-'||gen_random_uuid()::text,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),$11) RETURNING id::text", c.Key, customer, c.Customer.Name, c.Customer.Phone, c.Fulfillment.Type, address, c.Fulfillment.PickupAt, int64(math.Round(quote.Subtotal*100)), int64(math.Round(quote.DeliveryFee*100)), int64(math.Round(quote.Total*100)), hash).Scan(&id)
	return id, e
}
func (r Repository) InsertItems(ctx context.Context, tx shared.Transaction, id string, items []orders.PricedItem) error {
	for _, p := range items {
		_, e := transaction.Executor(r.DB, tx).Exec(ctx, "INSERT INTO app_schema.order_items(order_id,product_id,product_name_snapshot,unit_price_satang,quantity,line_total_satang) VALUES($1,$2,$3,$4,$5,$6)", id, p.ID, p.Name, p.Price, p.Quantity, p.Price*int64(p.Quantity))
		if e != nil {
			return e
		}
	}
	return nil
}
func (r Repository) GetTx(ctx context.Context, tx shared.Transaction, id string) (orders.Data, error) {
	return get(ctx, transaction.Executor(r.DB, tx), id)
}
func (r Repository) LockStatus(ctx context.Context, tx shared.Transaction, id string) (string, error) {
	var state string
	e := transaction.Executor(r.DB, tx).QueryRow(ctx, "SELECT status FROM app_schema.orders WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", id).Scan(&state)
	return state, missing(e)
}
func (r Repository) ChangeStatus(ctx context.Context, tx shared.Transaction, id, state, actor string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "UPDATE app_schema.orders SET status=$2,updated_by=NULLIF($3,'')::uuid WHERE id=$1", id, state, actor)
	return e
}
func (r Repository) History(ctx context.Context, tx shared.Transaction, id, from, to, actor, note string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "INSERT INTO app_schema.order_status_history(order_id,from_status,to_status,changed_by,note) VALUES($1,NULLIF($2,''),$3,NULLIF($4,'')::uuid,$5)", id, from, to, actor, note)
	return e
}
func (r Repository) ItemCount(ctx context.Context, tx shared.Transaction, id string) (int, error) {
	var n int
	e := transaction.Executor(r.DB, tx).QueryRow(ctx, "SELECT count(*) FROM app_schema.order_items WHERE order_id=$1 AND deleted_at IS NULL", id).Scan(&n)
	return n, e
}
