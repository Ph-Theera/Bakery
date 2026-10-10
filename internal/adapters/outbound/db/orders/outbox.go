package ordersdb

import (
	"bakery/internal/usecases/composite"
	"context"
)

func (r Repository) Pending(ctx context.Context) ([]composite.Message, error) {
	rows, err := r.DB.Query(ctx, "SELECT id::text,payload FROM app_schema.notification_outbox WHERE sent_at IS NULL AND next_attempt_at<=now() ORDER BY created_at LIMIT 20")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []composite.Message{}
	for rows.Next() {
		var m composite.Message
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
