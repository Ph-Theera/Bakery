package rest

import (
	"bakery/internal/domains/orders"
	orderin "bakery/internal/ports/port_in/orders"
	"bakery/internal/shared"
	"io"
	"net/http"
	"strings"
)

func (s Server) registerPayments(m *http.ServeMux) {
	paymentID := func(w http.ResponseWriter, r *http.Request) (string, string, bool) {
		id := r.PathValue("id")
		if !shared.UUID(id) {
			fail(w, shared.ErrInvalid)
			return "", "", false
		}
		order, e := s.Usecases.GetPaymentOrder.Handle(r.Context(), orderin.GetPaymentOrderQuery{PaymentID: id})
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
		v, e := s.Usecases.UploadPaymentSlip.Handle(r.Context(), orderin.UploadPaymentSlipCommand{PaymentID: id, OrderID: order, Mime: mime, Content: b})
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
		if e := s.Usecases.SubmitPayment.Handle(r.Context(), orderin.SubmitPaymentCommand{PaymentID: id, OrderID: order}); e != nil {
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
		v, e := s.Usecases.GetPaymentSlip.Handle(r.Context(), orderin.GetPaymentSlipQuery{PaymentID: id})
		if e != nil {
			fail(w, e)
			return
		}
		w.Header().Set("Content-Type", v.Mime)
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = w.Write(v.Content)
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
			var v orders.Data
			var e error
			if status == "approved" {
				v, e = s.Usecases.ApprovePayment.Handle(r.Context(), orderin.ApprovePaymentCommand{PaymentID: id, Actor: s.Auth.AdminID})
			} else {
				v, e = s.Usecases.RejectPayment.Handle(r.Context(), orderin.RejectPaymentCommand{PaymentID: id, Actor: s.Auth.AdminID, Reason: reason})
			}
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, v)
		}))
	}
}
