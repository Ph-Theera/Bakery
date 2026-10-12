package rest

import (
	"bakery/internal/domains/orders"
	orderin "bakery/internal/ports/port_in/orders"
	"bakery/internal/shared"
	"net/http"
)

func (s Server) registerOrders(m *http.ServeMux) {
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
		v, e := s.Usecases.GetOrder.Handle(r.Context(), orderin.GetOrderQuery{ID: id})
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
			v, e := s.Usecases.GetOrder.Handle(r.Context(), orderin.GetOrderQuery{ID: claims.OrderID})
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
		v, e := s.Usecases.ListMyOrders.Handle(r.Context(), orderin.ListMyOrdersQuery{UserID: claims.UserID, Limit: l, Offset: o})
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
		v, e := s.Usecases.ListOrders.Handle(r.Context(), orderin.ListOrdersQuery{Status: r.URL.Query().Get("status"), Limit: l, Offset: o})
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
		v, e := s.Usecases.UpdateOrderStatus.Handle(r.Context(), orderin.UpdateOrderStatusCommand{ID: id, Status: b.Status, Actor: s.Auth.AdminID, Note: b.Note})
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
		v, e := s.Usecases.UpdateOrderNote.Handle(r.Context(), orderin.UpdateOrderNoteCommand{ID: id, Note: b.Note, Actor: s.Auth.AdminID})
		if e != nil {
			fail(w, e)
			return
		}
		write(w, 200, v)
	}))
}
