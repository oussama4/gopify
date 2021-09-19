package gopify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// shopify API client
type Client struct {
	domain      string
	accessToken string
	version     string
	client      *http.Client
}

func NewClient(domain, accessToken, version string) *Client {
	client := http.Client{
		Timeout: time.Second * 10,
	}
	api := &Client{
		domain:      domain,
		accessToken: accessToken,
		version:     version,
		client:      &client,
	}

	return api
}

func (c *Client) Request(method string, url string, body io.Reader) (map[string]interface{}, error) {
	b, err := []byte(nil), error(nil)
	if body != nil {
		b, err = json.Marshal(&body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Shopify-Access-Token", c.accessToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	r := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r, nil
}

func (c *Client) Get(path string) (map[string]interface{}, error) {
	resourceUrl := fmt.Sprintf("https://%s/admin/api/%s/%s", c.domain, c.version, path)
	r, err := c.Request(http.MethodGet, resourceUrl, nil)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c *Client) Post(path string) (map[string]interface{}, error) {
	resourceUrl := fmt.Sprintf("https://%s/admin/api/%s/%s", c.domain, c.version, path)
	r, err := c.Request(http.MethodPost, resourceUrl, nil)
	if err != nil {
		return nil, err
	}
	return r, nil
}
