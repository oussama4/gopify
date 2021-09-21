package gopify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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
func (g *Gopify) VerifyRequest(u url.URL) error {
	q := u.Query()
	mac, err := hex.DecodeString(q.Get("hmac"))
	if err != nil {
		return err
	}
	q.Del("hmac")
	message, err := url.QueryUnescape(q.Encode())
	if err != nil {
		return err
	}

	hasher := hmac.New(sha256.New, []byte(g.ApiSecret))
	hasher.Write([]byte(message))
	validMac := hasher.Sum(nil)
	if !hmac.Equal(mac, validMac) {
		return ErrUnauthorizedRequest
	}
	return nil
}
