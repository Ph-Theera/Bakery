package productdb

import (
	"bakery/internal/domains/product"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMapError(t *testing.T) {
	for _, tc := range []struct{ in, want error }{{pgx.ErrNoRows, product.ErrNotFound}, {&pgconn.PgError{Code: "23505"}, product.ErrDuplicate}, {&pgconn.PgError{Code: "23514"}, product.ErrInvalid}} {
		if !errors.Is(mapError(tc.in), tc.want) {
			t.Fatal(tc.in)
		}
	}
	e := errors.New("network failure")
	if !errors.Is(mapError(e), e) {
		t.Fatal("error lost")
	}
}
func TestPostgresCRUD(t *testing.T) {
	uri := os.Getenv("TEST_DATABASE_URL")
	if uri == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cfg, e := pgxpool.ParseConfig(uri)
	if e != nil {
		t.Fatal("invalid test database URL")
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	db, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		t.Fatal("cannot open test database")
	}
	defer db.Close()
	r := New(db)
	sku := fmt.Sprintf("test-%d", time.Now().UnixNano())
	in := product.Input{SKU: sku, Name: "Bread", PriceSatang: 6500, Stock: 20}
	p, e := r.Create(ctx, in)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		cleanup, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = r.Delete(cleanup, p.ID)
	}()
	if _, e = r.Create(ctx, in); !errors.Is(e, product.ErrDuplicate) {
		t.Fatal(e)
	}
	got, e := r.Get(ctx, p.ID)
	if e != nil || got != p {
		t.Fatalf("get: %v %v", got, e)
	}
	in.Stock = 0
	got, e = r.Update(ctx, p.ID, in)
	if e != nil || got.Stock != 0 {
		t.Fatalf("update: %v %v", got, e)
	}
	ps, e := r.List(ctx, 100, 0)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, v := range ps {
		if v.ID == p.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("test row missing from first page; use empty test database")
	}
	if e = r.Delete(ctx, p.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = r.Get(ctx, p.ID); !errors.Is(e, product.ErrNotFound) {
		t.Fatal(e)
	}
	if _, e = r.Update(ctx, p.ID, in); !errors.Is(e, product.ErrNotFound) {
		t.Fatal(e)
	}
	if e = r.Delete(ctx, p.ID); !errors.Is(e, product.ErrNotFound) {
		t.Fatal(e)
	}
}
