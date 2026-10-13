package catalog

import (
	"bakery/internal/adapters/outbound/db/transaction"
	product "bakery/internal/domains/catalog"
	portout "bakery/internal/ports/port_out/catalog"
	"bakery/internal/shared"
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
func (r Repository) images(ctx context.Context, tx shared.Transaction, p *product.Data) error {
	return r.imagesWithQuery(ctx, transaction.Executor(r.DB, tx), p)
}
func (r Repository) imagesWithQuery(ctx context.Context, q transaction.Queryer, p *product.Data) error {
	rows, err := q.Query(ctx, "SELECT id::text,object_key,COALESCE(alt_text,''),is_primary,content IS NOT NULL FROM app_schema.product_images WHERE product_id=$1 AND deleted_at IS NULL ORDER BY is_primary DESC,sort_order,id", p.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var im product.Image
		var id string
		var stored bool
		if err = rows.Scan(&id, &im.PublicURL, &im.AltText, &im.Primary, &stored); err != nil {
			return err
		}
		if stored {
			im.PublicURL = "/api/v1/products/images/" + id
		} else {
			im.PublicURL = strings.TrimRight(r.ImageBase, "/") + "/" + strings.TrimLeft(im.PublicURL, "/")
		}
		p.Images = append(p.Images, im)
	}
	return rows.Err()
}

func (r Repository) AddImage(ctx context.Context, productID, mime, alt string, content []byte) (product.Data, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return product.Data{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var found string
	if err = tx.QueryRow(ctx, "SELECT id::text FROM app_schema.products WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", productID).Scan(&found); err != nil {
		return product.Data{}, dbError(err)
	}
	if _, err = tx.Exec(ctx, "UPDATE app_schema.product_images SET is_primary=false WHERE product_id=$1 AND is_primary AND deleted_at IS NULL", productID); err != nil {
		return product.Data{}, dbError(err)
	}
	if _, err = tx.Exec(ctx, "INSERT INTO app_schema.product_images(product_id,object_key,mime_type,alt_text,is_primary,content) VALUES($1,'db/'||gen_random_uuid()::text,$2,$3,true,$4)", productID, mime, alt, content); err != nil {
		return product.Data{}, dbError(err)
	}
	p, err := scan(tx.QueryRow(ctx, "SELECT "+cols+" FROM app_schema.products p WHERE p.id=$1", productID))
	if err != nil {
		return product.Data{}, err
	}
	if err = r.imagesWithQuery(ctx, tx, &p); err != nil {
		return product.Data{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return product.Data{}, err
	}
	return p, nil
}

func (r Repository) GetImage(ctx context.Context, id string) (string, []byte, error) {
	var mime string
	var content []byte
	err := r.DB.QueryRow(ctx, "SELECT i.mime_type,i.content FROM app_schema.product_images i JOIN app_schema.products p ON p.id=i.product_id WHERE i.id=$1 AND i.deleted_at IS NULL AND p.deleted_at IS NULL AND i.content IS NOT NULL", id).Scan(&mime, &content)
	return mime, content, dbError(err)
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
		if err = r.images(ctx, nil, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}
func (r Repository) Get(ctx context.Context, id string, public bool) (product.Data, error) {
	p, err := scan(r.DB.QueryRow(ctx, "SELECT "+cols+" FROM app_schema.products p WHERE (p.id::text=$1 OR p.slug=$1) AND p.deleted_at IS NULL AND (NOT $2 OR p.is_active)", id, public))
	if err == nil {
		err = r.images(ctx, nil, &p)
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
func NewProductRepository(db *pgxpool.Pool, base string) portout.Repository {
	return &Repository{DB: db, ImageBase: base}
}
func NewProductImageRepository(db *pgxpool.Pool, base string) portout.ImageRepository {
	return &Repository{DB: db, ImageBase: base}
}
func (r Repository) Save(ctx context.Context, tx shared.Transaction, id string, p product.Data) (product.Data, error) {
	q := transaction.Executor(r.DB, tx)
	var compare *int64
	if p.ComparePrice != nil {
		v := int64(math.Round(*p.ComparePrice * 100))
		compare = &v
	}
	var result string
	var e error
	if id == "" {
		e = q.QueryRow(ctx, "INSERT INTO app_schema.products(category_id,name,slug,description_th,price_satang,compare_at_price_satang,stock_quantity,is_active,is_featured) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id::text", p.CategoryID, p.Name, p.Slug, p.Description, int64(math.Round(p.Price*100)), compare, p.Stock, p.Active, p.Featured).Scan(&result)
	} else {
		e = q.QueryRow(ctx, "UPDATE app_schema.products SET category_id=$2,name=$3,slug=$4,description_th=$5,price_satang=$6,compare_at_price_satang=$7,stock_quantity=$8,is_active=$9,is_featured=$10 WHERE id=$1 AND deleted_at IS NULL RETURNING id::text", id, p.CategoryID, p.Name, p.Slug, p.Description, int64(math.Round(p.Price*100)), compare, p.Stock, p.Active, p.Featured).Scan(&result)
	}
	if e != nil {
		return product.Data{}, dbError(e)
	}
	out, e := scan(q.QueryRow(ctx, "SELECT "+cols+" FROM app_schema.products p WHERE p.id=$1", result))
	if e == nil {
		e = r.images(ctx, tx, &out)
	}
	return out, e
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
