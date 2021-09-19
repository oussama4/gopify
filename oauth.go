package gopify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// AuthorizationUrl returns a URL to shopify's consent page that asks for permissions
// for the required scopes.
func (g *Gopify) AuthorizationUrl(shop string) string {
	state := uniqueToken(15)
	query := url.Values{
		"client_id":    {g.ApiKey},
		"redirect_uri": {g.RedirectUrl},
		"scope":        {strings.Join(g.Scopes, ",")},
		"state":        {state},
	}
	return fmt.Sprintf("https://%s/admin/oauth/authorize?%s", shop, query.Encode())
}

// AccessToken retrieves an access token from shopify authorization server
//
// code is The authorization code obtained by using an authorization server
func (g *Gopify) AccessToken(shop string, code string) (string, error) {
	accessTokenPath := "admin/oauth/access_token"
	accessTokenEndPoint := fmt.Sprintf("https://%s/%s", shop, accessTokenPath)
	requestParams, err := json.Marshal(map[string]string{
		"client_id":     g.ApiKey,
		"client_secret": g.ApiSecret,
		"code":          code,
	})
	if err != nil {
		return "", nil
	}

	res, err := http.Post(accessTokenEndPoint, "application/json", bytes.NewBuffer(requestParams))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	resPayload := map[string]string{}
	if err := json.NewDecoder(res.Body).Decode(&resPayload); err != nil {
		return "", err
	}

	return resPayload["access_token"], nil
}
