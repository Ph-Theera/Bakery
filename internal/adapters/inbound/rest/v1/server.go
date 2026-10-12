package rest

import (
	orderin "bakery/internal/ports/port_in/orders"
	"bakery/internal/shared"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

type Server struct {
	Usecases                  *Usecases
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
		yes, err := s.Usecases.CheckOrderAccess.Handle(ctx, orderin.CheckOrderAccessQuery{ID: id, UserID: c.UserID})
		return err == nil && yes
	}
	return false
}

func NewRestServer(u *Usecases, auth Auth, lineSecret, origin string) Server {
	return Server{Usecases: u, Auth: auth, LineSecret: lineSecret, AllowedOrigin: origin}
}
func (s Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	s.registerStore(m)
	s.registerCatalog(m)
	s.registerCheckout(m)
	s.registerOrders(m)
	s.registerPayments(m)
	s.registerReporting(m)
	s.registerUsers(m)
	s.registerWebhook(m)
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
