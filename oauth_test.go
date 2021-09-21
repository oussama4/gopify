package gopify

import "testing"

func TestAuthorizationUrl(t *testing.T) {
	gopify := Gopify{
		ApiKey:      "key",
		ApiSecret:   "secret",
		RedirectUrl: "https://example.com/auth",
		Scopes:      []string{"read_products"},
	}
	cases := []struct {
		shop        string
		expectedUrl string
	}{
		{"osama.myshopify.com", "https://osama.myshopify.com/admin/oauth/authorize?client_id=key&redirect_uri=https%3A%2F%2Fexample.com%2Fauth&scope=read_products&state=state"},
	}

	for _, c := range cases {
		resultUrl := gopify.AuthorizationUrl(c.shop, "state")
		if resultUrl != c.expectedUrl {
			t.Errorf("gopify.AuthorizationUrl():\n got %s; \n want %s", resultUrl, c.expectedUrl)
		}
	}
}
