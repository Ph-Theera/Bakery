package api

import (
	"bakery/internal/shared"
	"bakery/internal/usecases/composite"
	"bakery/internal/usecases/orders"
	"bakery/internal/usecases/product"
	"bakery/internal/usecases/store"
	"bakery/internal/usecases/users"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	Catalog                   product.Service
	Orders                    orders.Repository
	Store                     store.Repository
	Users                     users.Repository
	Identity                  users.Identity
	Auth                      Auth
	LineSecret, AllowedOrigin string
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if status != 204 {
		_ = json.NewEncoder(w).Encode(v)
	}
}
func fail(w http.ResponseWriter, err error) {
	status, code := 500, "internal_error"
	switch {
	case errors.Is(err, shared.ErrInvalid):
		status, code = 400, "invalid_input"
	case errors.Is(err, shared.ErrNotFound):
		status, code = 404, "not_found"
	case errors.Is(err, shared.ErrConflict):
		status, code = 409, "conflict"
	case errors.Is(err, shared.ErrForbidden):
		status, code = 403, "forbidden"
	}
	write(w, status, map[string]any{"error": map[string]string{"code": code, "message": code}})
}
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil || d.Decode(new(any)) != io.EOF {
		return shared.ErrInvalid
	}
	return nil
}
func query(r *http.Request) (int, int, error) {
	l, o := 20, 0
	var err error
	if v := r.URL.Query().Get("limit"); v != "" {
		l, err = strconv.Atoi(v)
		if err != nil {
			return 0, 0, shared.ErrInvalid
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		o, err = strconv.Atoi(v)
		if err != nil {
			return 0, 0, shared.ErrInvalid
		}
	}
	if l < 1 || l > 100 || o < 0 {
		return 0, 0, shared.ErrInvalid
	}
	return l, o, nil
}
func (s Server) admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.Auth.Admin(r) {
			fail(w, shared.ErrForbidden)
			return
		}
		next(w, r)
	}
}
func (s Server) orderAccess(ctx context.Context, r *http.Request, id string) bool {
	if s.Auth.Admin(r) {
		return true
	}
	c, ok := s.Auth.Read(r)
	if !ok {
		return false
	}
	if c.OrderID == id {
		return true
	}
	if shared.UUID(c.UserID) {
		yes, err := s.Orders.Owned(ctx, id, c.UserID)
		return err == nil && yes
	}
	return false
}
func (s Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	m.HandleFunc("GET /store", func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Store.Get(r.Context())
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	})
	m.HandleFunc("GET /categories", func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Catalog.Repo.Categories(r.Context())
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
			v, e := s.Catalog.List(r.Context(), product.Query{Category: r.URL.Query().Get("category"), Search: r.URL.Query().Get("search"), ActiveOnly: public, Limit: l, Offset: o})
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
		v, e := s.Catalog.Get(r.Context(), r.PathValue("slug"), true)
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
		id := r.PathValue("id")
		v, e := s.Catalog.Save(r.Context(), id, p)
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
		p, e := s.Catalog.Get(r.Context(), id, false)
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
		v, e := s.Catalog.Save(r.Context(), id, p)
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
		if e := s.Catalog.Repo.Delete(r.Context(), id, s.Auth.AdminID); e != nil {
			fail(w, e)
			return
		}
		write(w, 204, nil)
	}))
	checkout := composite.Checkout{Repo: s.Orders}
	m.HandleFunc("POST /checkout/quote", func(w http.ResponseWriter, r *http.Request) {
		var c orders.Command
		if e := decode(w, r, &c); e != nil {
			fail(w, e)
			return
		}
		v, e := checkout.Quote(r.Context(), c)
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	})
	m.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
		if len(s.Auth.Secret) < 32 {
			write(w, 503, map[string]string{"error": "AUTH_SECRET not configured"})
			return
		}
		var c orders.Command
		if e := decode(w, r, &c); e != nil {
			fail(w, e)
			return
		}
		c.Key = r.Header.Get("Idempotency-Key")
		if claims, ok := s.Auth.Read(r); ok {
			c.UserID = claims.UserID
		}
		v, e := checkout.Create(r.Context(), c)
		if e != nil {
			fail(w, e)
			return
		}
		v.AccessToken = s.Auth.Sign(Claims{OrderID: v.ID, Expires: time.Now().Add(7 * 24 * time.Hour).Unix()})
		write(w, 201, v)
	})
	m.HandleFunc("GET /orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !shared.UUID(id) {
			fail(w, shared.ErrInvalid)
			return
		}
		if !s.orderAccess(r.Context(), r, id) {
			fail(w, shared.ErrForbidden)
			return
		}
		v, e := s.Orders.Get(r.Context(), id)
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	})
	m.HandleFunc("GET /orders", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := s.Auth.Read(r)
		if !ok {
			fail(w, shared.ErrForbidden)
			return
		}
		if claims.OrderID != "" {
			v, e := s.Orders.Get(r.Context(), claims.OrderID)
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, []orders.Data{v})
			return
		}
		l, o, e := query(r)
		if e != nil {
			fail(w, e)
			return
		}
		if !shared.UUID(claims.UserID) {
			fail(w, shared.ErrForbidden)
			return
		}
		v, e := s.Orders.Mine(r.Context(), claims.UserID, l, o)
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	})
	m.HandleFunc("GET /admin/orders", s.admin(func(w http.ResponseWriter, r *http.Request) {
		l, o, e := query(r)
		if e != nil {
			fail(w, e)
			return
		}
		v, e := s.Orders.List(r.Context(), r.URL.Query().Get("status"), l, o)
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
	m.HandleFunc("PATCH /admin/orders/{id}/status", s.admin(func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Status string `json:"status"`
			Note   string `json:"note"`
		}
		if e := decode(w, r, &b); e != nil {
			fail(w, e)
			return
		}
		id := r.PathValue("id")
		if !shared.UUID(id) {
			fail(w, shared.ErrInvalid)
			return
		}
		v, e := s.Orders.Status(r.Context(), id, b.Status, s.Auth.AdminID, b.Note)
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
	m.HandleFunc("PATCH /admin/orders/{id}", s.admin(func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Note string `json:"internal_note"`
		}
		if e := decode(w, r, &b); e != nil {
			fail(w, e)
			return
		}
		id := r.PathValue("id")
		if !shared.UUID(id) || len(b.Note) > 5000 {
			fail(w, shared.ErrInvalid)
			return
		}
		v, e := s.Orders.Note(r.Context(), id, b.Note, s.Auth.AdminID)
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
	paymentID := func(w http.ResponseWriter, r *http.Request) (string, string, bool) {
		id := r.PathValue("id")
		if !shared.UUID(id) {
			fail(w, shared.ErrInvalid)
			return "", "", false
		}
		order, e := s.Orders.PaymentOrder(r.Context(), id)
		if e != nil {
			fail(w, e)
			return "", "", false
		}
		if !s.orderAccess(r.Context(), r, order) {
			fail(w, shared.ErrForbidden)
			return "", "", false
		}
		return id, order, true
	}
	m.HandleFunc("POST /payments/{id}/slip", func(w http.ResponseWriter, r *http.Request) {
		id, order, ok := paymentID(w, r)
		if !ok {
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, (5<<20)+(64<<10))
		if e := r.ParseMultipartForm(5 << 20); e != nil {
			fail(w, shared.ErrInvalid)
			return
		}
		if r.MultipartForm != nil {
			defer func() { _ = r.MultipartForm.RemoveAll() }()
		}
		f, _, e := r.FormFile("slip")
		if e != nil {
			fail(w, shared.ErrInvalid)
			return
		}
		defer func() { _ = f.Close() }()
		b, e := io.ReadAll(io.LimitReader(f, (5<<20)+1))
		if e != nil || len(b) == 0 || len(b) > 5<<20 {
			fail(w, shared.ErrInvalid)
			return
		}
		mime := http.DetectContentType(b)
		if mime != "image/jpeg" && mime != "image/png" && mime != "image/webp" {
			fail(w, shared.ErrInvalid)
			return
		}
		v, e := s.Orders.Slip(r.Context(), id, order, mime, b)
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 201, v)
	})
	m.HandleFunc("POST /payments/{id}/submit", func(w http.ResponseWriter, r *http.Request) {
		id, order, ok := paymentID(w, r)
		if !ok {
			return
		}
		if e := s.Orders.Submit(r.Context(), id, order); e != nil {
			fail(w, e)
			return
		}
		write(w, 200, orders.Evidence{ID: id, Status: "submitted"})
	})
	m.HandleFunc("GET /admin/payments/{id}/slip", s.admin(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !shared.UUID(id) {
			fail(w, shared.ErrInvalid)
			return
		}
		b, mime, e := s.Orders.Evidence(r.Context(), id)
		if e != nil {
			fail(w, e)
			return
		}
		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = w.Write(b)
	}))
	for _, action := range []string{"approve", "reject"} {
		action := action
		m.HandleFunc("POST /admin/payments/{id}/"+action, s.admin(func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			if !shared.UUID(id) {
				fail(w, shared.ErrInvalid)
				return
			}
			status, reason := "approved", ""
			if action == "reject" {
				var b struct {
					Reason string `json:"reason"`
				}
				if e := decode(w, r, &b); e != nil {
					fail(w, e)
					return
				}
				reason = strings.TrimSpace(b.Reason)
				if reason == "" || len(reason) > 2000 {
					fail(w, shared.ErrInvalid)
					return
				}
				status = "rejected"
			}
			v, e := s.Orders.Review(r.Context(), id, status, s.Auth.AdminID, reason)
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, v)
		}))
	}
	m.HandleFunc("GET /admin/dashboard", s.admin(func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Orders.Dashboard(r.Context())
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
	m.HandleFunc("GET /admin/customers", s.admin(func(w http.ResponseWriter, r *http.Request) {
		l, o, e := query(r)
		if e != nil {
			fail(w, e)
			return
		}
		v, e := s.Users.Customers(r.Context(), l, o)
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
	m.HandleFunc("GET /admin/settings", s.admin(func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Store.Get(r.Context())
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
	m.HandleFunc("PATCH /admin/settings", s.admin(func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Store.Get(r.Context())
		if e != nil {
			fail(w, e)
			return
		}
		if e = decode(w, r, &v); e != nil {
			fail(w, e)
			return
		}
		if len(v.Name) == 0 || len(v.Name) > 200 || len(v.Description) > 5000 || math.IsNaN(v.Fee) || math.IsInf(v.Fee, 0) || v.Fee < 0 || v.Fee > 1e6 || v.FreeThreshold < 0 || v.FreeThreshold > 1e9 {
			fail(w, shared.ErrInvalid)
			return
		}
		v, e = s.Store.Save(r.Context(), v)
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
	m.HandleFunc("POST /users/line/link", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Token string `json:"access_token"`
		}
		if e := decode(w, r, &b); e != nil {
			fail(w, e)
			return
		}
		if len(s.Auth.Secret) < 32 {
			fail(w, shared.ErrForbidden)
			return
		}
		p, e := s.Identity.Verify(r.Context(), b.Token)
		if e != nil {
			fail(w, e)
			return
		}
		p, e = s.Users.Link(r.Context(), p)
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, map[string]any{"user": p, "access_token": s.Auth.Sign(Claims{UserID: p.ID, Expires: time.Now().Add(time.Hour).Unix()})})
	})
	m.HandleFunc("GET /users/me", func(w http.ResponseWriter, r *http.Request) {
		c, ok := s.Auth.Read(r)
		if !ok || !shared.UUID(c.UserID) {
			fail(w, shared.ErrForbidden)
			return
		}
		write(w, 200, map[string]string{"id": c.UserID})
	})
	m.HandleFunc("POST /webhooks/line", func(w http.ResponseWriter, r *http.Request) {
		if s.LineSecret == "" {
			fail(w, shared.ErrForbidden)
			return
		}
		raw, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 2<<20))
		if e != nil {
			fail(w, shared.ErrInvalid)
			return
		}
		sig, e := base64.StdEncoding.DecodeString(r.Header.Get("X-Line-Signature"))
		mac := hmac.New(sha256.New, []byte(s.LineSecret))
		_, _ = mac.Write(raw)
		if e != nil || !hmac.Equal(sig, mac.Sum(nil)) {
			fail(w, shared.ErrForbidden)
			return
		}
		if !json.Valid(raw) {
			fail(w, shared.ErrInvalid)
			return
		}
		write(w, 200, map[string]any{"accepted": true})
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if origin := r.Header.Get("Origin"); origin != "" && origin == s.AllowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,Idempotency-Key")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		m.ServeHTTP(w, r)
	})
}
