package gopify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
)

const (
	ShopifyHmacHeader = "X-Shopify-Hmac-SHA256"
)

var (
	ErrInvalidWebhookRequest = errors.New("webhook request is unauthorized")
)

// VerifyWebhook verifies that webhook request is from shopify
func (g *Gopify) VerifyWebhook(r *http.Request) bool {
	mac := r.Header.Get(ShopifyHmacHeader)
	body, _ := ioutil.ReadAll(r.Body)

	hasher := hmac.New(sha256.New, []byte(g.ApiSecret))
	hasher.Write(body)
	validMac := hex.EncodeToString(hasher.Sum(nil))
	validMac = base64.StdEncoding.EncodeToString([]byte(validMac))

	return hmac.Equal([]byte(validMac), []byte(mac))
}
