package gopify

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifyWebhook(t *testing.T) {
	gopify := Gopify{
		ApiKey:      "key",
		ApiSecret:   "hush",
		RedirectUrl: "https://example.com/auth",
		Scopes:      []string{"read_products"},
	}

	cases := []struct {
		payload  []byte
		mac      string
		expected bool
	}{
		{[]byte("webhook request body"), "MzE4OWFmOThjYmIyODA2ZmZmZWFmMjdmYzQ2ZTg3MTM1M2FmZTNlYmMzMGYzNTNkMDA0ZjQyNjkxMGZjNzEzNA==", true},
		{[]byte("webhook request body"), "wronghash", false},
	}

	for _, c := range cases {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(c.payload))
		req.Header.Add("X-Shopify-Hmac-SHA256", c.mac)
		valid := gopify.VerifyWebhook(req)

		if valid != c.expected {
			t.Errorf("webhook verification expected %v got %v", c.expected, valid)
		}
	}
}
