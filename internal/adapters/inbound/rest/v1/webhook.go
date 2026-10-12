package rest

import (
	"bakery/internal/shared"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
)

func (s Server) registerWebhook(m *http.ServeMux) {
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
}
