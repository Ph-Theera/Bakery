package notificationsdb

import (
	"bakery/internal/adapters/outbound/db/transaction"
	"bakery/internal/domains/notifications"
	portout "bakery/internal/ports/port_out/notifications"
	"bakery/internal/shared"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func (r Repository) Pending(ctx context.Context) ([]notifications.Message, error) {
	rows, err := r.DB.Query(ctx, "SELECT id::text,payload FROM app_schema.notification_outbox WHERE sent_at IS NULL AND next_attempt_at<=now() ORDER BY created_at LIMIT 20")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []notifications.Message{}
	for rows.Next() {
		var m notifications.Message
		if err = rows.Scan(&m.ID, &m.Text); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (r Repository) Finish(ctx context.Context, id string, sent bool) error {
	_, err := r.DB.Exec(ctx, "UPDATE app_schema.notification_outbox SET attempts=attempts+1,next_attempt_at=now()+interval '1 minute'*least(attempts+1,60),sent_at=CASE WHEN $2 THEN now() ELSE sent_at END WHERE id=$1", id, sent)
	return err
}

type Repository struct{ DB *pgxpool.Pool }

func NewOutboxRepository(db *pgxpool.Pool) portout.Outbox { return &Repository{DB: db} }
func (r Repository) Enqueue(ctx context.Context, tx shared.Transaction, id, text string) error {
	_, e := transaction.Executor(r.DB, tx).Exec(ctx, "INSERT INTO app_schema.notification_outbox(order_id,payload) VALUES($1,$2) ON CONFLICT(order_id) DO NOTHING", id, text)
	return e
}
