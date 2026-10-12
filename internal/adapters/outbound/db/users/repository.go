package usersdb

import (
	"bakery/internal/adapters/outbound/db/transaction"
	"bakery/internal/domains/users"
	portout "bakery/internal/ports/port_out/users"
	"bakery/internal/shared"
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
func NewUserRepository(db *pgxpool.Pool) portout.Repository { return &Repository{DB: db} }
func (r Repository) UpsertCustomer(ctx context.Context, tx shared.Transaction, name, phone, user string) (string, error) {
	var id string
	e := transaction.Executor(r.DB, tx).QueryRow(ctx, "INSERT INTO app_schema.customers(display_name,phone,user_id) VALUES($1,$2,NULLIF($3,'')::uuid) ON CONFLICT(user_id) DO UPDATE SET display_name=EXCLUDED.display_name,phone=EXCLUDED.phone RETURNING id::text", name, phone, user).Scan(&id)
	return id, e
}
