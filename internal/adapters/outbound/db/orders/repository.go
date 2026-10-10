package ordersdb

import (
	"bakery/internal/shared"
	"bakery/internal/usecases/orders"
	"bakery/internal/usecases/store"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ DB *pgxpool.Pool }
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

type priced struct {
	ID, Name string
	Quantity int
	Price    int64
}

func price(ctx context.Context, q querier, c orders.Command, lock bool) (orders.Quote, []priced, error) {
	var raw []byte
	if err := q.QueryRow(ctx, "SELECT value FROM app_schema.store_settings WHERE key='store'").Scan(&raw); err != nil {
		return orders.Quote{}, nil, err
	}
	var s store.Settings
	if err := json.Unmarshal(raw, &s); err != nil {
		return orders.Quote{}, nil, err
	}
	if !s.Accepting {
		return orders.Quote{}, nil, shared.ErrConflict
	}
	items := append([]orders.Item{}, c.Items...)
	sort.Slice(items, func(i, j int) bool { return items[i].ProductID < items[j].ProductID })
	out := []priced{}
	var subtotal int64
	for _, i := range items {
		var p priced
		var stock int64
		p.ID = i.ProductID
		p.Quantity = i.Quantity
		sql := "SELECT name,price_satang,stock_quantity FROM app_schema.products WHERE id=$1 AND is_active AND deleted_at IS NULL"
		if lock {
			sql += " FOR UPDATE"
		}
		if err := q.QueryRow(ctx, sql, i.ProductID).Scan(&p.Name, &p.Price, &stock); err != nil {
			return orders.Quote{}, nil, missing(err)
		}
		var reserved int64
		if err := q.QueryRow(ctx, "SELECT COALESCE(sum(quantity),0) FROM app_schema.inventory_reservations WHERE product_id=$1 AND status='reserved' AND expires_at>now() AND deleted_at IS NULL", i.ProductID).Scan(&reserved); err != nil {
			return orders.Quote{}, nil, err
		}
		if int64(i.Quantity) > stock-reserved {
			return orders.Quote{}, nil, shared.ErrConflict
		}
		if p.Price < 0 || p.Price > 1e11 {
			return orders.Quote{}, nil, shared.ErrInvalid
		}
		subtotal += p.Price * int64(i.Quantity)
		out = append(out, p)
	}
	fee := int64(math.Round(s.Fee * 100))
	if c.Fulfillment.Type == "pickup" || s.FreeThreshold > 0 && subtotal >= int64(math.Round(s.FreeThreshold*100)) {
		fee = 0
	}
	return orders.Quote{Subtotal: float64(subtotal) / 100, DeliveryFee: float64(fee) / 100, Total: float64(subtotal+fee) / 100}, out, nil
}
func (r Repository) Quote(ctx context.Context, c orders.Command) (orders.Quote, error) {
	q, _, err := price(ctx, r.DB, c, false)
	return q, err
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
func (r Repository) Create(ctx context.Context, c orders.Command) (orders.Data, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return orders.Data{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", c.Key); err != nil {
		return orders.Data{}, err
	}
	raw, err := json.Marshal(struct {
		Command orders.Command
		UserID  string
	}{c, c.UserID})
	if err != nil {
		return orders.Data{}, err
	}
	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])
	var existing, oldHash string
	err = tx.QueryRow(ctx, "SELECT id::text,COALESCE(request_hash,'') FROM app_schema.orders WHERE idempotency_key=$1", c.Key).Scan(&existing, &oldHash)
	if err == nil {
		if oldHash != hash {
			return orders.Data{}, shared.ErrConflict
		}
		return get(ctx, tx, existing)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return orders.Data{}, err
	}
	quote, items, err := price(ctx, tx, c, true)
	if err != nil {
		return orders.Data{}, err
	}
	var customer string
	err = tx.QueryRow(ctx, "INSERT INTO app_schema.customers(display_name,phone,user_id) VALUES($1,$2,NULLIF($3,'')::uuid) ON CONFLICT(user_id) DO UPDATE SET display_name=EXCLUDED.display_name,phone=EXCLUDED.phone RETURNING id::text", c.Customer.Name, c.Customer.Phone, c.UserID).Scan(&customer)
	if err != nil {
		return orders.Data{}, err
	}
	var address any
	if c.Fulfillment.Type == "delivery" {
		b, _ := json.Marshal(map[string]string{"address": c.Fulfillment.Address})
		// Exec mode has no inferred JSONB OID; []byte would encode as bytea.
		address = string(b)
	}
	var id string
	err = tx.QueryRow(ctx, "INSERT INTO app_schema.orders(order_number,idempotency_key,customer_id,customer_name_snapshot,customer_phone_snapshot,fulfillment_type,delivery_address_snapshot,pickup_at,subtotal_satang,delivery_fee_satang,total_satang,placed_at,request_hash) VALUES('TB-'||gen_random_uuid()::text,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),$11) RETURNING id::text", c.Key, customer, c.Customer.Name, c.Customer.Phone, c.Fulfillment.Type, address, c.Fulfillment.PickupAt, int64(math.Round(quote.Subtotal*100)), int64(math.Round(quote.DeliveryFee*100)), int64(math.Round(quote.Total*100)), hash).Scan(&id)
	if err != nil {
		return orders.Data{}, err
	}
	for _, p := range items {
		_, err = tx.Exec(ctx, "INSERT INTO app_schema.order_items(order_id,product_id,product_name_snapshot,unit_price_satang,quantity,line_total_satang) VALUES($1,$2,$3,$4,$5,$6)", id, p.ID, p.Name, p.Price, p.Quantity, p.Price*int64(p.Quantity))
		if err != nil {
			return orders.Data{}, err
		}
		_, err = tx.Exec(ctx, "INSERT INTO app_schema.inventory_reservations(order_id,product_id,quantity,expires_at) VALUES($1,$2,$3,now()+interval '30 minutes')", id, p.ID, p.Quantity)
		if err != nil {
			return orders.Data{}, err
		}
	}
	_, err = tx.Exec(ctx, "INSERT INTO app_schema.payments(order_id,method,expected_amount_satang,idempotency_key) VALUES($1,$2,$3,$4)", id, c.PaymentMethod, int64(math.Round(quote.Total*100)), c.Key)
	if err != nil {
		return orders.Data{}, err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO app_schema.order_status_history(order_id,to_status) VALUES($1,'pending_payment')", id); err != nil {
		return orders.Data{}, err
	}
	o, err := get(ctx, tx, id)
	if err != nil {
		return o, err
	}
	return o, tx.Commit(ctx)
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
func history(ctx context.Context, tx pgx.Tx, id, from, to, actor, note string) error {
	_, err := tx.Exec(ctx, "INSERT INTO app_schema.order_status_history(order_id,from_status,to_status,changed_by,note) VALUES($1,$2,$3,NULLIF($4,'')::uuid,$5)", id, from, to, actor, note)
	return err
}
func (r Repository) Status(ctx context.Context, id, status, actor, note string) (orders.Data, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return orders.Data{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var old string
	err = tx.QueryRow(ctx, "SELECT status FROM app_schema.orders WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", id).Scan(&old)
	if err != nil {
		return orders.Data{}, missing(err)
	}
	next := map[string][]string{"pending_payment": {"cancelled"}, "pending_verification": {"cancelled"}, "confirmed": {"preparing"}, "preparing": {"ready_for_pickup", "shipped"}, "ready_for_pickup": {"completed"}, "shipped": {"completed"}}
	allowed := old == status
	for _, v := range next[old] {
		if v == status {
			allowed = true
		}
	}
	if !allowed {
		return orders.Data{}, shared.ErrConflict
	}
	if status == "cancelled" {
		_, err = tx.Exec(ctx, "UPDATE app_schema.inventory_reservations SET status='released',released_at=now() WHERE order_id=$1 AND status='reserved'", id)
		if err != nil {
			return orders.Data{}, err
		}
		_, err = tx.Exec(ctx, "UPDATE app_schema.payments SET status='expired' WHERE order_id=$1 AND status IN ('pending','submitted','rejected')", id)
		if err != nil {
			return orders.Data{}, err
		}
	}
	_, err = tx.Exec(ctx, "UPDATE app_schema.orders SET status=$2,updated_by=$3 WHERE id=$1", id, status, actor)
	if err != nil {
		return orders.Data{}, err
	}
	if old != status {
		if err = history(ctx, tx, id, old, status, actor, note); err != nil {
			return orders.Data{}, err
		}
	}
	o, err := get(ctx, tx, id)
	if err != nil {
		return o, err
	}
	return o, tx.Commit(ctx)
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
func (r Repository) Slip(ctx context.Context, payment, order, mime string, content []byte) (orders.Evidence, error) {
	sum := sha256.Sum256(content)
	hash := hex.EncodeToString(sum[:])
	var id string
	err := r.DB.QueryRow(ctx, "INSERT INTO app_schema.payment_evidence(payment_id,object_key,mime_type,file_size_bytes,checksum,content) SELECT id,'private/'||$3,$4,$5,$3,$6 FROM app_schema.payments WHERE id=$1 AND order_id=$2 AND status IN ('pending','rejected','needs_resubmission') AND deleted_at IS NULL RETURNING id::text", payment, order, hash, mime, len(content), content).Scan(&id)
	return orders.Evidence{ID: id, Status: "uploaded"}, missing(err)
}
func (r Repository) Submit(ctx context.Context, payment, order string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var state string
	err = tx.QueryRow(ctx, "SELECT status FROM app_schema.orders WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", order).Scan(&state)
	if err != nil {
		return missing(err)
	}
	if state != "pending_payment" && state != "pending_verification" {
		return shared.ErrConflict
	}
	tag, err := tx.Exec(ctx, "UPDATE app_schema.payments SET status='submitted' WHERE id=$1 AND order_id=$2 AND status IN ('pending','rejected','needs_resubmission','submitted') AND EXISTS (SELECT 1 FROM app_schema.payment_evidence WHERE payment_id=$1 AND deleted_at IS NULL) AND EXISTS (SELECT 1 FROM app_schema.inventory_reservations WHERE order_id=$2 AND status='reserved' AND expires_at>now())", payment, order)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrConflict
	}
	_, err = tx.Exec(ctx, "UPDATE app_schema.orders SET status='pending_verification' WHERE id=$1", order)
	if err != nil {
		return err
	}
	if state != "pending_verification" {
		if err = history(ctx, tx, order, state, "pending_verification", "", ""); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (r Repository) Review(ctx context.Context, payment, status, actor, reason string) (orders.Data, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return orders.Data{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id string
	err = tx.QueryRow(ctx, "SELECT order_id::text FROM app_schema.payments WHERE id=$1 AND deleted_at IS NULL", payment).Scan(&id)
	if err != nil {
		return orders.Data{}, missing(err)
	}
	var old string
	err = tx.QueryRow(ctx, "SELECT status FROM app_schema.orders WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", id).Scan(&old)
	if err != nil {
		return orders.Data{}, missing(err)
	}
	var pstate string
	err = tx.QueryRow(ctx, "SELECT status FROM app_schema.payments WHERE id=$1 FOR UPDATE", payment).Scan(&pstate)
	if err != nil {
		return orders.Data{}, err
	}
	if pstate == status {
		return get(ctx, tx, id)
	}
	if pstate != "submitted" || old != "pending_verification" {
		return orders.Data{}, shared.ErrConflict
	}
	if status == "approved" {
		rows, e := tx.Query(ctx, "SELECT product_id::text,quantity FROM app_schema.inventory_reservations WHERE order_id=$1 AND status='reserved' AND expires_at>now() ORDER BY product_id", id)
		if e != nil {
			return orders.Data{}, e
		}
		items := []orders.Item{}
		for rows.Next() {
			var i orders.Item
			if e = rows.Scan(&i.ProductID, &i.Quantity); e != nil {
				rows.Close()
				return orders.Data{}, e
			}
			items = append(items, i)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return orders.Data{}, e
		}
		if len(items) == 0 {
			return orders.Data{}, shared.ErrConflict
		}
		var count int
		if e = tx.QueryRow(ctx, "SELECT count(*) FROM app_schema.order_items WHERE order_id=$1", id).Scan(&count); e != nil {
			return orders.Data{}, e
		}
		if count != len(items) {
			return orders.Data{}, shared.ErrConflict
		}
		for _, i := range items {
			tag, e := tx.Exec(ctx, "UPDATE app_schema.products SET stock_quantity=stock_quantity-$2 WHERE id=$1 AND stock_quantity>=$2", i.ProductID, i.Quantity)
			if e != nil {
				return orders.Data{}, e
			}
			if tag.RowsAffected() != 1 {
				return orders.Data{}, shared.ErrConflict
			}
			_, e = tx.Exec(ctx, "INSERT INTO app_schema.inventory_movements(product_id,quantity_delta,reason,reference_type,reference_id,idempotency_key,created_by) VALUES($1,$2,'sale','order',$3,$4,$5)", i.ProductID, -i.Quantity, id, id+":"+i.ProductID, actor)
			if e != nil {
				return orders.Data{}, e
			}
		}
		_, err = tx.Exec(ctx, "UPDATE app_schema.inventory_reservations SET status='committed' WHERE order_id=$1 AND status='reserved'", id)
		if err != nil {
			return orders.Data{}, err
		}
		_, err = tx.Exec(ctx, "UPDATE app_schema.payments SET status='approved',paid_amount_satang=expected_amount_satang,verified_by=$2,verified_at=now() WHERE id=$1", payment, actor)
		if err != nil {
			return orders.Data{}, err
		}
		_, err = tx.Exec(ctx, "UPDATE app_schema.orders SET status='confirmed' WHERE id=$1", id)
		if err != nil {
			return orders.Data{}, err
		}
		if err = history(ctx, tx, id, old, "confirmed", actor, ""); err != nil {
			return orders.Data{}, err
		}
	} else {
		_, err = tx.Exec(ctx, "UPDATE app_schema.payments SET status='rejected',rejection_reason=$2,verified_by=$3,verified_at=now() WHERE id=$1", payment, reason, actor)
		if err != nil {
			return orders.Data{}, err
		}
		_, err = tx.Exec(ctx, "UPDATE app_schema.orders SET status='pending_payment' WHERE id=$1", id)
		if err != nil {
			return orders.Data{}, err
		}
		if err = history(ctx, tx, id, old, "pending_payment", actor, reason); err != nil {
			return orders.Data{}, err
		}
	}
	o, err := get(ctx, tx, id)
	if err != nil {
		return o, err
	}
	if status == "approved" {
		_, err = tx.Exec(ctx, "INSERT INTO app_schema.notification_outbox(order_id,payload) VALUES($1,$2) ON CONFLICT(order_id) DO NOTHING", id, fmt.Sprintf("TongBai: อนุมัติการชำระเงินแล้ว\nออเดอร์ %s\nลูกค้า %s\nยอด %.2f บาท", o.Number, o.Customer, o.Total))
		if err != nil {
			return o, err
		}
	}
	return o, tx.Commit(ctx)
}
func (r Repository) Dashboard(ctx context.Context) (map[string]any, error) {
	var sales int64
	var count, pending int
	err := r.DB.QueryRow(ctx, "SELECT COALESCE(sum(total_satang) FILTER (WHERE status IN ('confirmed','preparing','ready_for_pickup','shipped','completed')),0),count(*),count(*) FILTER (WHERE status='pending_verification') FROM app_schema.orders WHERE deleted_at IS NULL").Scan(&sales, &count, &pending)
	return map[string]any{"total_sales_thb": float64(sales) / 100, "total_orders": count, "pending_payments": pending}, err
}
func (r Repository) PaymentOrder(ctx context.Context, payment string) (string, error) {
	var id string
	err := r.DB.QueryRow(ctx, "SELECT order_id::text FROM app_schema.payments WHERE id=$1 AND deleted_at IS NULL", payment).Scan(&id)
	return id, missing(err)
}
func (r Repository) Evidence(ctx context.Context, payment string) ([]byte, string, error) {
	var b []byte
	var mime string
	err := r.DB.QueryRow(ctx, "SELECT content,mime_type FROM app_schema.payment_evidence WHERE payment_id=$1 AND deleted_at IS NULL AND content IS NOT NULL ORDER BY created_at DESC,id LIMIT 1", payment).Scan(&b, &mime)
	return b, mime, missing(err)
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
