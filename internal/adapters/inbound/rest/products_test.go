package rest

import (
	"bakery/internal/domains/product"
	"bakery/internal/ports/port_out"
	"bakery/internal/usecases"
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

type stub struct {
	port_out.ProductRepository
	err error
}

func (s stub) Create(_ context.Context, i product.Input) (product.Data, error) {
	return product.Data{ID: 1, SKU: i.SKU, Name: i.Name, PriceSatang: i.PriceSatang, Stock: i.Stock}, s.err
}
func (s stub) Get(context.Context, int64) (product.Data, error)       { return product.Data{}, s.err }
func (s stub) List(context.Context, int, int) ([]product.Data, error) { return []product.Data{}, s.err }
func TestHTTP(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body string
		err                      error
		status                   int
	}{
		{"create", "POST", "/products", `{"sku":"A","name":"Bread","price_satang":0,"stock":0}`, nil, 201},
		{"missing price", "POST", "/products", `{"sku":"A","name":"Bread","stock":0}`, nil, 400},
		{"null price", "POST", "/products", `{"sku":"A","name":"Bread","price_satang":null,"stock":0}`, nil, 400},
		{"trailing JSON", "POST", "/products", `{"sku":"A","name":"Bread","price_satang":0,"stock":0} {}`, nil, 400},
		{"unknown field", "POST", "/products", `{"extra":true}`, nil, 400},
		{"duplicate", "POST", "/products", `{"sku":"A","name":"Bread","price_satang":0,"stock":0}`, product.ErrDuplicate, 409},
		{"not found", "GET", "/products/1", "", product.ErrNotFound, 404},
		{"invalid id", "GET", "/products/no", "", nil, 400},
		{"invalid limit", "GET", "/products?limit=101", "", nil, 400},
		{"empty list", "GET", "/products", "", nil, 200},
		{"masked error", "GET", "/products/1", "", errors.New("secret password SQL"), 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := stub{err: tc.err}
			h := New(usecases.NewCreateProduct(r), usecases.NewGetProduct(r), usecases.NewListProducts(r), usecases.NewUpdateProduct(r), usecases.NewDeleteProduct(r))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			if w.Code != tc.status {
				t.Fatalf("status %d body %s", w.Code, w.Body.String())
			}
			if strings.Contains(w.Body.String(), "secret") {
				t.Fatal("leaked error")
			}
			if tc.name == "empty list" && strings.TrimSpace(w.Body.String()) != "[]" {
				t.Fatal(w.Body.String())
			}
			if tc.name == "create" && w.Header().Get("Location") != "/products/1" {
				t.Fatal("missing location")
			}
		})
	}
}
