package rest

import (
	storein "bakery/internal/ports/port_in/store"
	"net/http"
)

func (s Server) registerStore(m *http.ServeMux) {
	m.HandleFunc("GET /store", func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Usecases.GetStore.Handle(r.Context(), storein.GetStoreQuery{})
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	})
	m.HandleFunc("GET /admin/settings", s.admin(func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Usecases.GetStore.Handle(r.Context(), storein.GetStoreQuery{})
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
	m.HandleFunc("PATCH /admin/settings", s.admin(func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Usecases.GetStore.Handle(r.Context(), storein.GetStoreQuery{})
		if e != nil {
			fail(w, e)
			return
		}
		if e = decode(w, r, &v); e != nil {
			fail(w, e)
			return
		}
		v, e = s.Usecases.UpdateStore.Handle(r.Context(), storein.UpdateStoreCommand{Settings: v})
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
}
