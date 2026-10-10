package rest

import (
	"bakery/internal/domains/product"
	portin "bakery/internal/ports/port_in/products"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

type request struct {
	SKU   string `json:"sku"`
	Name  string `json:"name"`
	Price *int64 `json:"price_satang"`
	Stock *int64 `json:"stock"`
}
type response struct {
	ID    int64  `json:"id"`
	SKU   string `json:"sku"`
	Name  string `json:"name"`
	Price int64  `json:"price_satang"`
	Stock int64  `json:"stock"`
}

func dto(p product.Data) response { return response{p.ID, p.SKU, p.Name, p.PriceSatang, p.Stock} }
func write(w http.ResponseWriter, s int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, e error) {
	s, m := 500, "internal server error"
	switch {
	case errors.Is(e, product.ErrInvalid):
		s, m = 400, "invalid product or query"
	case errors.Is(e, product.ErrNotFound):
		s, m = 404, "product not found"
	case errors.Is(e, product.ErrDuplicate):
		s, m = 409, "sku already exists"
	}
	write(w, s, map[string]string{"error": m})
}
func decode(w http.ResponseWriter, r *http.Request) (product.Input, error) {
	var b request
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	d.DisallowUnknownFields()
	if d.Decode(&b) != nil {
		return product.Input{}, product.ErrInvalid
	}
	if d.Decode(new(any)) != io.EOF || b.Price == nil || b.Stock == nil {
		return product.Input{}, product.ErrInvalid
	}
	return product.Input{SKU: b.SKU, Name: b.Name, PriceSatang: *b.Price, Stock: *b.Stock}, nil
}
func id(r *http.Request) (int64, error) {
	n, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil || n <= 0 {
		return 0, product.ErrInvalid
	}
	return n, nil
}
func New(c portin.CreateProductUsecase, g portin.GetProductUsecase, l portin.ListProductsUsecase, u portin.UpdateProductUsecase, d portin.DeleteProductUsecase) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /products", func(w http.ResponseWriter, r *http.Request) {
		in, e := decode(w, r)
		if e != nil {
			fail(w, e)
			return
		}
		p, e := c.Handle(r.Context(), in)
		if e != nil {
			fail(w, e)
			return
		}
		w.Header().Set("Location", "/products/"+strconv.FormatInt(p.ID, 10))
		write(w, 201, dto(p))
	})
	mux.HandleFunc("GET /products/{id}", func(w http.ResponseWriter, r *http.Request) {
		n, e := id(r)
		if e != nil {
			fail(w, e)
			return
		}
		p, e := g.Handle(r.Context(), n)
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, dto(p))
	})
	mux.HandleFunc("PUT /products/{id}", func(w http.ResponseWriter, r *http.Request) {
		n, e := id(r)
		if e != nil {
			fail(w, e)
			return
		}
		in, e := decode(w, r)
		if e != nil {
			fail(w, e)
			return
		}
		p, e := u.Handle(r.Context(), portin.UpdateProductCommand{ID: n, Input: in})
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, dto(p))
	})
	mux.HandleFunc("DELETE /products/{id}", func(w http.ResponseWriter, r *http.Request) {
		n, e := id(r)
		if e != nil {
			fail(w, e)
			return
		}
		if e = d.Handle(r.Context(), n); e != nil {
			fail(w, e)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("GET /products", func(w http.ResponseWriter, r *http.Request) {
		q := portin.ListProductsQuery{Limit: 20}
		var e error
		if v, ok := r.URL.Query()["limit"]; ok {
			q.Limit, e = strconv.Atoi(v[0])
			if e != nil {
				fail(w, product.ErrInvalid)
				return
			}
		}
		if v, ok := r.URL.Query()["offset"]; ok {
			q.Offset, e = strconv.Atoi(v[0])
			if e != nil {
				fail(w, product.ErrInvalid)
				return
			}
		}
		ps, e := l.Handle(r.Context(), q)
		if e != nil {
			fail(w, e)
			return
		}
		out := make([]response, 0, len(ps))
		for _, p := range ps {
			out = append(out, dto(p))
		}
		write(w, 200, out)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}
