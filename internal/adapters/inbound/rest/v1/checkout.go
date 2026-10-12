package rest

import (
	"bakery/internal/domains/orders"
	"net/http"
	"time"
)

func (s Server) registerCheckout(m *http.ServeMux) {
	m.HandleFunc("POST /checkout/quote", func(w http.ResponseWriter, r *http.Request) {
		var c orders.Command
		if e := decode(w, r, &c); e != nil {
			fail(w, e)
			return
		}
		v, e := s.Usecases.QuoteCheckout.Handle(r.Context(), c)
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
		v, e := s.Usecases.CreateOrder.Handle(r.Context(), c)
		if e != nil {
			fail(w, e)
			return
		}
		v.AccessToken = s.Auth.Sign(Claims{OrderID: v.ID, Expires: time.Now().Add(7 * 24 * time.Hour).Unix()})
		write(w, 201, v)
	})
}
