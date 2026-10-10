package catalog

import (
	"bakery/internal/shared"
	"bakery/internal/usecases/product"
	"context"
	"errors"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB        *pgxpool.Pool
	ImageBase string
}

func dbError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.ErrNotFound
	}
	var e *pgconn.PgError
	if errors.As(err, &e) {
		switch e.Code {
		case "23505":
			return shared.ErrConflict
		case "23503", "23514", "22001", "22P02":
			return shared.ErrInvalid
		}
	}
	return err
}

const cols = "p.id::text,p.category_id::text,p.name,COALESCE(p.slug,''),p.description_th,p.price_satang,p.compare_at_price_satang,p.stock_quantity,p.is_active,p.is_featured"

func scan(row pgx.Row) (product.Data, error) {
	var p product.Data
	var price int64
	var compare *int64
	err := row.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Slug, &p.Description, &price, &compare, &p.Stock, &p.Active, &p.Featured)
	p.Price = float64(price) / 100
	if compare != nil {
		v := float64(*compare) / 100
		p.ComparePrice = &v
	}
	p.Images = []product.Image{}
	return p, dbError(err)
}
func (r Repository) images(ctx context.Context, p *product.Data) error {
	rows, err := r.DB.Query(ctx, "SELECT object_key,COALESCE(alt_text,''),is_primary FROM app_schema.product_images WHERE product_id=$1 AND deleted_at IS NULL ORDER BY is_primary DESC,sort_order,id", p.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var im product.Image
		if err = rows.Scan(&im.PublicURL, &im.AltText, &im.Primary); err != nil {
			return err
		}
		im.PublicURL = strings.TrimRight(r.ImageBase, "/") + "/" + strings.TrimLeft(im.PublicURL, "/")
		p.Images = append(p.Images, im)
	}
	return rows.Err()
}
func (r Repository) List(ctx context.Context, q product.Query) ([]product.Data, error) {
	rows, err := r.DB.Query(ctx, "SELECT "+cols+" FROM app_schema.products p WHERE p.deleted_at IS NULL AND (NOT $1 OR p.is_active) AND ($2='' OR p.category_id::text=$2) AND ($3='' OR p.name ILIKE '%'||$3||'%') ORDER BY p.created_at DESC,p.id LIMIT $4 OFFSET $5", q.ActiveOnly, q.Category, q.Search, q.Limit, q.Offset)
	if err != nil {
		return nil, dbError(err)
	}
	out := []product.Data{}
	for rows.Next() {
		p, e := scan(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, p)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for i := range out {
		if err = r.images(ctx, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}
func (r Repository) Get(ctx context.Context, id string, public bool) (product.Data, error) {
	p, err := scan(r.DB.QueryRow(ctx, "SELECT "+cols+" FROM app_schema.products p WHERE (p.id::text=$1 OR p.slug=$1) AND p.deleted_at IS NULL AND (NOT $2 OR p.is_active)", id, public))
	if err == nil {
		err = r.images(ctx, &p)
	}
	return p, err
}
func (r Repository) Categories(ctx context.Context) ([]product.Category, error) {
	rows, err := r.DB.Query(ctx, "SELECT id::text,COALESCE(slug,''),COALESCE(name_th,'') FROM app_schema.categories WHERE deleted_at IS NULL AND is_active ORDER BY sort_order,id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []product.Category{}
	for rows.Next() {
		var c product.Category
		if err = rows.Scan(&c.ID, &c.Slug, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (r Repository) Save(ctx context.Context, id string, p product.Data) (product.Data, error) {
	var compare *int64
	if p.ComparePrice != nil {
		v := int64(math.Round(*p.ComparePrice * 100))
		compare = &v
	}
	var result string
	var err error
	if id == "" {
		err = r.DB.QueryRow(ctx, "INSERT INTO app_schema.products(category_id,name,slug,description_th,price_satang,compare_at_price_satang,stock_quantity,is_active,is_featured) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id::text", p.CategoryID, p.Name, p.Slug, p.Description, int64(math.Round(p.Price*100)), compare, p.Stock, p.Active, p.Featured).Scan(&result)
	} else {
		tx, e := r.DB.Begin(ctx)
		if e != nil {
			return product.Data{}, e
		}
		defer func() { _ = tx.Rollback(ctx) }()
		var reserved int64
		var locked string
		if e = tx.QueryRow(ctx, "SELECT id::text FROM app_schema.products WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", id).Scan(&locked); e != nil {
			return product.Data{}, dbError(e)
		}
		if e = tx.QueryRow(ctx, "SELECT COALESCE(sum(quantity),0) FROM app_schema.inventory_reservations WHERE product_id=$1 AND status='reserved' AND expires_at>now()", id).Scan(&reserved); e != nil {
			return product.Data{}, e
		}
		if int64(p.Stock) < reserved {
			return product.Data{}, shared.ErrConflict
		}
		err = tx.QueryRow(ctx, "UPDATE app_schema.products SET category_id=$2,name=$3,slug=$4,description_th=$5,price_satang=$6,compare_at_price_satang=$7,stock_quantity=$8,is_active=$9,is_featured=$10 WHERE id=$1 AND deleted_at IS NULL RETURNING id::text", id, p.CategoryID, p.Name, p.Slug, p.Description, int64(math.Round(p.Price*100)), compare, p.Stock, p.Active, p.Featured).Scan(&result)
		if err != nil {
			return product.Data{}, dbError(err)
		}
		if err = tx.Commit(ctx); err != nil {
			return product.Data{}, err
		}
	}
	if err != nil {
		return product.Data{}, dbError(err)
	}
	return r.Get(ctx, result, false)
}
func (r Repository) Delete(ctx context.Context, id, actor string) error {
	tag, err := r.DB.Exec(ctx, "UPDATE app_schema.products SET deleted_at=now(),deleted_by=$2,is_active=false WHERE id=$1 AND deleted_at IS NULL", id, actor)
	if err != nil {
		return dbError(err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}
