package gopify

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifyRequest(t *testing.T) {
	gopify := Gopify{
		ApiKey:      "key",
		ApiSecret:   "hush",
		RedirectUrl: "https://example.com/auth",
		Scopes:      []string{"read_products"},
	}

	cases := []struct {
		u        string
		expected int
	}{
		{"code=0907a61c0c8d55e99db179b68161bc00&hmac=700e2dadb827fcc8609e9d5ce208b2e9cdaab9df07390d2cbca10d7c328fc4bf&shop=some-shop.myshopify.com&state=0.6784241404160823&timestamp=1337178173", http.StatusOK},
		{"code=0907a61c0c8d55e99db179b68161bc00&hmac=700e2dadb827fcc8609e9d5ce208b2e9cdaab9df07390d2cbca10d7c328fc4bf&shop=some-shop.myshopify.com&state=0.6784241404160823&timestamp=133717817", http.StatusUnauthorized},
	}

	mux := http.DefaultServeMux
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	mux.Handle("/", gopify.VerifyRequest(h))

	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", c.u), nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		res := rec.Result()

		if res.StatusCode != c.expected {
			t.Errorf("incorrect response, got %d", res.StatusCode)
		}
	}
}
