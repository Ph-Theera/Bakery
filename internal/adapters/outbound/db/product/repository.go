package productdb

import (
	"bakery/internal/domains/product"
	portout "bakery/internal/ports/port_out/products"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) portout.ProductRepository { return &Repository{db} }

const columns = "id,sku,name,price_satang,stock"

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return product.ErrNotFound
	}
	var e *pgconn.PgError
	if errors.As(err, &e) {
		switch e.Code {
		case "23505":
			return product.ErrDuplicate
		case "23514", "22001":
			return product.ErrInvalid
		}
	}
	return err
}
func scan(row pgx.Row) (product.Data, error) {
	var p product.Data
	err := row.Scan(&p.ID, &p.SKU, &p.Name, &p.PriceSatang, &p.Stock)
	return p, mapError(err)
}
func (r *Repository) Create(ctx context.Context, i product.Input) (product.Data, error) {
	return scan(r.db.QueryRow(ctx, "INSERT INTO public.products(sku,name,price_satang,stock) VALUES($1,$2,$3,$4) RETURNING "+columns, i.SKU, i.Name, i.PriceSatang, i.Stock))
}
func (r *Repository) Get(ctx context.Context, id int64) (product.Data, error) {
	return scan(r.db.QueryRow(ctx, "SELECT "+columns+" FROM public.products WHERE id=$1", id))
}
func (r *Repository) Update(ctx context.Context, id int64, i product.Input) (product.Data, error) {
	return scan(r.db.QueryRow(ctx, "UPDATE public.products SET sku=$2,name=$3,price_satang=$4,stock=$5 WHERE id=$1 RETURNING "+columns, id, i.SKU, i.Name, i.PriceSatang, i.Stock))
}
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM public.products WHERE id=$1", id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return product.ErrNotFound
	}
	return nil
}
func (r *Repository) List(ctx context.Context, limit, offset int) ([]product.Data, error) {
	rows, err := r.db.Query(ctx, "SELECT "+columns+" FROM public.products ORDER BY id LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	out := make([]product.Data, 0)
	for rows.Next() {
		p, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, mapError(rows.Err())
}
