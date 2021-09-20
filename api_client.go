package gopify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// shopify API client
type Client struct {
	domain      string
	baseUrl     string
	accessToken string
	version     string
	client      *http.Client
}

func NewClient(domain, accessToken, version string) *Client {
	baseUrl := fmt.Sprintf("https://%s/admin/api/%s", domain, version)
	client := http.Client{
		Timeout: time.Second * 10,
	}
	api := &Client{
		domain:      domain,
		baseUrl:     baseUrl,
		accessToken: accessToken,
		version:     version,
		client:      &client,
	}

	return api
}

func (c *Client) Request(method string, path string, body map[string]interface{}) (map[string]interface{}, error) {
	u := fmt.Sprintf("%s/%s", c.baseUrl, path)
	b, err := []byte(nil), error(nil)
	if body != nil {
		b, err = json.Marshal(&body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u, bytes.NewBuffer(b))
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
	r, err := c.Request(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c *Client) Post(path string, body map[string]interface{}) (map[string]interface{}, error) {
	r, err := c.Request(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c *Client) Put(path string, body map[string]interface{}) (map[string]interface{}, error) {
	r, err := c.Request(http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}
	return r, err
}

func (c *Client) Delete(path string) error {
	_, err := c.Request(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return nil
}
