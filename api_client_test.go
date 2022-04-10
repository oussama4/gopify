package gopify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Shopify-Access-Token")
		if token != "valid access token" {
			failureResponse := map[string]any{
				"errors": "Invalid API key or access token",
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(failureResponse)
			return
		}
		successResponse := map[string]any{
			"products": []map[string]any{
				{
					"title": "Product 1",
				},
			},
		}
		json.NewEncoder(w).Encode(successResponse)
	}))
	defer ts.Close()

	cases := []struct {
		accessToken string
		want        map[string]any
		err         error
	}{
		{
			accessToken: "valid access token",
			want:        map[string]any{"products": []map[string]any{{"title": "Product 1"}}},
			err:         nil,
		},
		{
			accessToken: "invalid access token",
			want:        map[string]any{},
			err:         ResponseError{Errors: `"Invalid API key or access token"`},
		},
	}

	for _, c := range cases {
		apiClient := NewClient(ts.URL[7:], c.accessToken)
		apiClient.baseUrl = fmt.Sprintf("%s/admin/api/%s", ts.URL, apiClient.version)
		res := map[string]any{}
		_, err := apiClient.Get("products.json", nil, &res)
		if err != c.err {
			t.Errorf("Expected error %v, got %v", c.err, err)
		}
		if fmt.Sprint(res) != fmt.Sprint(c.want) {
			t.Errorf("Expected response %v, got %v", c.want, res)
		}
	}
}
