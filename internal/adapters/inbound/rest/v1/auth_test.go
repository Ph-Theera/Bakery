package rest

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSignedAccess(t *testing.T) {
	a := Auth{Secret: strings.Repeat("x", 32)}
	for _, tc := range []struct {
		name  string
		token string
		valid bool
	}{{"valid", a.Sign(Claims{OrderID: "order-1", Expires: time.Now().Add(time.Hour).Unix()}), true}, {"expired", a.Sign(Claims{Expires: time.Now().Add(-time.Minute).Unix()}), false}, {"forged", a.Sign(Claims{Expires: time.Now().Add(time.Hour).Unix()}) + "x", false}} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/orders", nil)
			r.Header.Set("Authorization", "Bearer "+tc.token)
			_, ok := a.Read(r)
			if ok != tc.valid {
				t.Fatal(ok)
			}
		})
	}
}
func TestAdminRoutesDenyUntrustedRequests(t *testing.T) {
	s := Server{Auth: Auth{AdminToken: "admin-secret", AdminID: "admin-id"}}
	for _, path := range []string{"/admin/products", "/admin/orders", "/admin/customers", "/admin/settings", "/admin/dashboard"} {
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 403 {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
}
