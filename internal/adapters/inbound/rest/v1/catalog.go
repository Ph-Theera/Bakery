package rest

import (
	product "bakery/internal/domains/catalog"
	catalogin "bakery/internal/ports/port_in/catalog"
	"bakery/internal/shared"
	"encoding/json"
	"net/http"
)

func (s Server) registerCatalog(m *http.ServeMux) {
	m.HandleFunc("GET /categories", func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Usecases.ListCategories.Handle(r.Context(), catalogin.ListCategoriesQuery{})
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	})
	listProducts := func(public bool) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			l, o, e := query(r)
			if e != nil {
				fail(w, e)
				return
			}
			v, e := s.Usecases.ListProducts.Handle(r.Context(), catalogin.ListProductsQuery{Category: r.URL.Query().Get("category"), Search: r.URL.Query().Get("search"), ActiveOnly: public, Limit: l, Offset: o})
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, v)
		}
	}
	m.HandleFunc("GET /products", listProducts(true))
	m.HandleFunc("GET /admin/products", s.admin(listProducts(false)))
	m.HandleFunc("GET /products/{slug}", func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Usecases.GetProduct.Handle(r.Context(), catalogin.GetProductQuery{ID: r.PathValue("slug"), Public: true})
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	})
	saveProduct := func(w http.ResponseWriter, r *http.Request) {
		var p product.Data
		if e := decode(w, r, &p); e != nil {
			fail(w, e)
			return
		}
		v, e := s.Usecases.CreateProduct.Handle(r.Context(), catalogin.CreateProductCommand{Product: p})
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 201, v)
	}
	m.HandleFunc("POST /admin/products", s.admin(saveProduct))
	m.HandleFunc("PATCH /admin/products/{id}", s.admin(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !shared.UUID(id) {
			fail(w, shared.ErrInvalid)
			return
		}
		p, e := s.Usecases.GetProduct.Handle(r.Context(), catalogin.GetProductQuery{ID: id, Public: false})
		if e != nil {
			fail(w, e)
			return
		}
		p.Images = nil
		var fields map[string]json.RawMessage
		if e = decode(w, r, &fields); e != nil {
			fail(w, e)
			return
		}
		allowed := map[string]bool{"category_id": true, "name": true, "slug": true, "description_th": true, "price_thb": true, "compare_at_price_thb": true, "stock_quantity": true, "is_active": true, "is_featured": true}
		for k, v := range fields {
			if !allowed[k] || string(v) == "null" && k != "compare_at_price_thb" {
				fail(w, shared.ErrInvalid)
				return
			}
		}
		raw, _ := json.Marshal(fields)
		if e = json.Unmarshal(raw, &p); e != nil {
			fail(w, shared.ErrInvalid)
			return
		}
		v, e := s.Usecases.UpdateProduct.Handle(r.Context(), catalogin.UpdateProductCommand{ID: id, Product: p})
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
	m.HandleFunc("DELETE /admin/products/{id}", s.admin(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !shared.UUID(id) {
			fail(w, shared.ErrInvalid)
			return
		}
		if e := s.Usecases.DeleteProduct.Handle(r.Context(), catalogin.DeleteProductCommand{ID: id, Actor: s.Auth.AdminID}); e != nil {
			fail(w, e)
			return
		}
		write(w, 204, nil)
	}))
}
