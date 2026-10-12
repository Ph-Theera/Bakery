package rest

import (
	userin "bakery/internal/ports/port_in/users"
	"bakery/internal/shared"
	"net/http"
	"time"
)

func (s Server) registerUsers(m *http.ServeMux) {
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
		p, e := s.Usecases.LinkLine.Handle(r.Context(), userin.LinkLineCommand{Token: b.Token})
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
}
