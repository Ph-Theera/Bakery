package usecases

import (
	"bakery/internal/domains/product"
	"bakery/internal/ports/port_in"
	"bakery/internal/ports/port_out"
	"context"
	"errors"
	"testing"
)

type fakeRepo struct {
	port_out.ProductRepository
	calls int
	err   error
	input product.Input
}

func (f *fakeRepo) Create(_ context.Context, i product.Input) (product.Data, error) {
	f.calls++
	f.input = i
	return product.Data{ID: 1, SKU: i.SKU}, f.err
}
func TestCreateProduct(t *testing.T) {
	sentinel := errors.New("database failure")
	for _, tc := range []struct {
		name               string
		in                 product.Input
		repoErr, errorWant error
		calls              int
	}{
		{"success", product.Input{SKU: " A ", Name: " Bread "}, nil, nil, 1},
		{"empty name", product.Input{SKU: "A"}, nil, product.ErrInvalid, 0},
		{"negative price", product.Input{SKU: "A", Name: "Bread", PriceSatang: -1}, nil, product.ErrInvalid, 0},
		{"negative stock", product.Input{SKU: "A", Name: "Bread", Stock: -1}, nil, product.ErrInvalid, 0},
		{"duplicate", product.Input{SKU: "A", Name: "Bread"}, product.ErrDuplicate, product.ErrDuplicate, 1},
		{"failure", product.Input{SKU: "A", Name: "Bread"}, sentinel, sentinel, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeRepo{err: tc.repoErr}
			_, e := NewCreateProduct(r).Handle(context.Background(), tc.in)
			if !errors.Is(e, tc.errorWant) || r.calls != tc.calls {
				t.Fatalf("error %v, calls %d", e, r.calls)
			}
			if tc.name == "success" && (r.input.SKU != "A" || r.input.Name != "Bread") {
				t.Fatal("input not normalized")
			}
		})
	}
}
func TestInvalidQueriesDoNotCallRepository(t *testing.T) {
	r := &fakeRepo{}
	ctx := context.Background()
	_, e := NewGetProduct(r).Handle(ctx, 0)
	if !errors.Is(e, product.ErrInvalid) {
		t.Fatal(e)
	}
	_, e = NewUpdateProduct(r).Handle(ctx, port_in.UpdateProductCommand{ID: -1})
	if !errors.Is(e, product.ErrInvalid) {
		t.Fatal(e)
	}
	if e = NewDeleteProduct(r).Handle(ctx, 0); !errors.Is(e, product.ErrInvalid) {
		t.Fatal(e)
	}
	for _, q := range []port_in.ListProductsQuery{{Limit: 0}, {Limit: 101}, {Limit: 20, Offset: -1}} {
		if _, e = NewListProducts(r).Handle(ctx, q); !errors.Is(e, product.ErrInvalid) {
			t.Fatal(e)
		}
	}
}
