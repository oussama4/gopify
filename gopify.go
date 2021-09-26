package gopify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
)

var (
	ErrUnauthorizedRequest = errors.New("unauthorized request")
)

// Gopify holds common shopify app settings
type Gopify struct {
	ApiKey      string
	ApiSecret   string
	RedirectUrl string
	Scopes      []string
}

// VerifyRequest verifies the authenticity of the request from Shopify
func (g *Gopify) VerifyRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		mac, err := hex.DecodeString(q.Get("hmac"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		q.Del("hmac")
		message, _ := url.QueryUnescape(q.Encode())

		if !g.ValidHmac(mac, message) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ValidHmac validates the provided hmac value against the hmac value of the provided message
func (g *Gopify) ValidHmac(mac []byte, message string) bool {
	hasher := hmac.New(sha256.New, []byte(g.ApiSecret))
	hasher.Write([]byte(message))
	validMac := hasher.Sum(nil)
	return hmac.Equal(mac, validMac)
}
