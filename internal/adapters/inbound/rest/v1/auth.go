package rest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Auth struct{ Secret, AdminToken, AdminID string }
type Claims struct {
	OrderID string `json:"order,omitempty"`
	UserID  string `json:"user,omitempty"`
	Expires int64  `json:"exp"`
}

func (a Auth) Sign(c Claims) string {
	if len(a.Secret) < 32 {
		return ""
	}
	raw, _ := json.Marshal(c)
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, []byte(a.Secret))
	_, _ = mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func (a Auth) Read(r *http.Request) (Claims, bool) {
	if len(a.Secret) < 32 || !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		return Claims{}, false
	}
	s := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	p := strings.Split(s, ".")
	if len(p) != 2 {
		return Claims{}, false
	}
	sig, err := base64.RawURLEncoding.DecodeString(p[1])
	if err != nil {
		return Claims{}, false
	}
	mac := hmac.New(sha256.New, []byte(a.Secret))
	_, _ = mac.Write([]byte(p[0]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return Claims{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(p[0])
	if err != nil {
		return Claims{}, false
	}
	var c Claims
	err = json.Unmarshal(raw, &c)
	return c, err == nil && c.Expires > time.Now().Unix()
}
func (a Auth) Admin(r *http.Request) bool {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	return strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") && a.AdminToken != "" && a.AdminID != "" && hmac.Equal([]byte(token), []byte(a.AdminToken))
}
