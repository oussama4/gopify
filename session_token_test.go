package gopify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// create a session token for testing purposes
func createToken(expired bool, key string, secret string) string {
	h := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	header, _ := json.Marshal(h)
	headerEncoded := base64.RawURLEncoding.EncodeToString(header)

	now := time.Now().Unix()
	exp := int(now) + 60
	if expired {
		exp = int(now) - 60
	}
	p := Payload{
		Iss:  "https://shop-name.myshopify.com/admin",
		Dest: "https://shop-name.myshopify.com",
		Aud:  key,
		Sub:  "userid",
		Exp:  exp,
		Nbf:  int(now),
		Iat:  int(now),
		Jti:  "f8912129-1af6-4cad-9ca3-76b0f7621087",
		Sid:  "aaea182f2732d44c23057c0fea584021a4485b2bd25d3eb7fd349313ad24c685",
	}
	payload, _ := json.Marshal(p)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payload)

	hasher := hmac.New(sha256.New, []byte(secret))
	hasher.Write([]byte(fmt.Sprintf("%s.%s", headerEncoded, payloadEncoded)))
	signature := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))

	return fmt.Sprintf("%s.%s.%s", headerEncoded, payloadEncoded, signature)
}

func TestVerifyToken(t *testing.T) {
	gopify := Gopify{
		ApiKey:      "key",
		ApiSecret:   "hush",
		RedirectUrl: "https://example.com/auth",
		Scopes:      []string{"read_products"},
	}
	validToken := createToken(false, gopify.ApiKey, gopify.ApiSecret)
	expiredToken := createToken(true, gopify.ApiKey, gopify.ApiSecret)
	invalidToken := createToken(false, "wrongKey", gopify.ApiSecret)
	invalidSignatureToken := createToken(false, gopify.ApiKey, "wrongSecret")
	cases := []struct {
		token    string
		expected int
	}{
		{validToken, http.StatusOK},
		{expiredToken, http.StatusUnauthorized},
		{"", http.StatusBadRequest},
		{invalidToken, http.StatusBadRequest},
		{invalidSignatureToken, http.StatusUnauthorized},
	}

	mux := http.DefaultServeMux
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	mux.Handle("/test/token", gopify.VerifyToken(h))

	for i, c := range cases {
		req := httptest.NewRequest(http.MethodGet, "/test/token", nil)
		authHeader := fmt.Sprintf("Bearer %s", c.token)
		req.Header.Add("Authorization", authHeader)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		res := rec.Result()

		if res.StatusCode != c.expected {
			t.Errorf("case %d expected %d status code but got %d", i, c.expected, res.StatusCode)
		}
	}
}
