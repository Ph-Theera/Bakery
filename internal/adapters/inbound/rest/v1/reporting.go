package rest

import (
	orderin "bakery/internal/ports/port_in/orders"
	userin "bakery/internal/ports/port_in/users"
	"net/http"
)

func (s Server) registerReporting(m *http.ServeMux) {
	m.HandleFunc("GET /admin/dashboard", s.admin(func(w http.ResponseWriter, r *http.Request) {
		v, e := s.Usecases.GetDashboard.Handle(r.Context(), orderin.GetDashboardQuery{})
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
		v, e := s.Usecases.ListCustomers.Handle(r.Context(), userin.ListCustomersQuery{Limit: l, Offset: o})
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
}
