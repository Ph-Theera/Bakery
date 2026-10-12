package paymentsdb

import (
	"bakery/internal/adapters/outbound/db/transaction"
	"bakery/internal/domains/orders"
	portout "bakery/internal/ports/port_out/payments"
	"bakery/internal/shared"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ DB *pgxpool.Pool }

func NewPaymentRepository(db *pgxpool.Pool) portout.Repository { return &Repository{DB: db} }
func missing(e error) error {
	if errors.Is(e, pgx.ErrNoRows) {
		return shared.ErrNotFound
	}
	return e
}
func (r Repository) Insert(ctx context.Context, tx shared.Transaction, id, method string, amount int64, key string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "INSERT INTO app_schema.payments(order_id,method,expected_amount_satang,idempotency_key) VALUES($1,$2,$3,$4)", id, method, amount, key)
	return e
}
func (r Repository) OrderID(ctx context.Context, id string) (string, error) {
	var order string
	e := r.DB.QueryRow(ctx, "SELECT order_id::text FROM app_schema.payments WHERE id=$1 AND deleted_at IS NULL", id).Scan(&order)
	return order, missing(e)
}
func (r Repository) LockStatus(ctx context.Context, tx shared.Transaction, id string) (string, error) {
	var state string
	e := transaction.Executor(r.DB, tx).QueryRow(ctx, "SELECT status FROM app_schema.payments WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", id).Scan(&state)
	return state, missing(e)
}
func (r Repository) HasEvidence(ctx context.Context, tx shared.Transaction, id string) (bool, error) {
	var yes bool
	e := transaction.Executor(r.DB, tx).QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM app_schema.payment_evidence WHERE payment_id=$1 AND deleted_at IS NULL)", id).Scan(&yes)
	return yes, e
}
func (r Repository) Slip(ctx context.Context, tx shared.Transaction, payment, mime string, content []byte) (orders.Evidence, error) {
	sum := sha256.Sum256(content)
	hash := hex.EncodeToString(sum[:])
	var id string
	e := transaction.Executor(r.DB, tx).QueryRow(ctx, "INSERT INTO app_schema.payment_evidence(payment_id,object_key,mime_type,file_size_bytes,checksum,content) VALUES($1,'private/'||$2,$3,$4,$2,$5) RETURNING id::text", payment, hash, mime, len(content), content).Scan(&id)
	return orders.Evidence{ID: id, Status: "uploaded"}, missing(e)
}
func (r Repository) Submit(ctx context.Context, tx shared.Transaction, id string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "UPDATE app_schema.payments SET status='submitted' WHERE id=$1", id)
	return e
}
func (r Repository) Approve(ctx context.Context, tx shared.Transaction, id, actor string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "UPDATE app_schema.payments SET status='approved',paid_amount_satang=expected_amount_satang,verified_by=$2,verified_at=now() WHERE id=$1", id, actor)
	return e
}
func (r Repository) Reject(ctx context.Context, tx shared.Transaction, id, actor, reason string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "UPDATE app_schema.payments SET status='rejected',rejection_reason=$2,verified_by=$3,verified_at=now() WHERE id=$1", id, reason, actor)
	return e
}
func (r Repository) CancelPending(ctx context.Context, tx shared.Transaction, order string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "UPDATE app_schema.payments SET status='expired' WHERE order_id=$1 AND status IN ('pending','submitted','rejected')", order)
	return e
}
func (r Repository) Evidence(ctx context.Context, payment string) ([]byte, string, error) {
	var b []byte
	var mime string
	e := r.DB.QueryRow(ctx, "SELECT content,mime_type FROM app_schema.payment_evidence WHERE payment_id=$1 AND deleted_at IS NULL AND content IS NOT NULL ORDER BY created_at DESC,id LIMIT 1", payment).Scan(&b, &mime)
	return b, mime, missing(e)
}
