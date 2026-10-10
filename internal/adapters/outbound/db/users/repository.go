package usersdb

import (
	"bakery/internal/usecases/users"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ DB *pgxpool.Pool }

func (r Repository) Customers(ctx context.Context, limit, offset int) ([]users.Customer, error) {
	rows, err := r.DB.Query(ctx, "SELECT c.id::text,c.display_name,c.phone,count(o.id) FROM app_schema.customers c LEFT JOIN app_schema.orders o ON o.customer_id=c.id AND o.deleted_at IS NULL WHERE c.deleted_at IS NULL GROUP BY c.id ORDER BY c.created_at DESC,c.id LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []users.Customer{}
	for rows.Next() {
		var c users.Customer
		if err = rows.Scan(&c.ID, &c.Name, &c.Phone, &c.OrderCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (r Repository) Link(ctx context.Context, p users.Profile) (users.Profile, error) {
	var id string
	err := r.DB.QueryRow(ctx, "INSERT INTO app_schema.users(auth_provider,provider_subject) VALUES('line',$1) ON CONFLICT(auth_provider,provider_subject) DO UPDATE SET updated_at=now() WHERE app_schema.users.is_active AND app_schema.users.deleted_at IS NULL RETURNING id::text", p.ID).Scan(&id)
	p.ID = id
	return p, err
}
