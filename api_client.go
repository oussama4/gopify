package gopify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultApiVersion = "2021-10"
	defaultTimeout    = 10 * time.Second
	defaultRetries    = 2
)

var (
	ErrRateLimit = errors.New("API rate limit exeeded")
)

// GraphqlErr is a general graphql error
type GraphqlError struct {
	msg string
}

func (err GraphqlError) Error() string {
	return err.msg
}

// Option is for configuring the Api client
type Option func(c *Client)

// WithVersion sets the Api version
func WithVersion(version string) Option {
	return func(c *Client) {
		c.version = version
	}
}

// WithTimeout sets a custom http timeout
func WithTimeout(seconds int) Option {
	return func(c *Client) {
		c.client.Timeout = time.Duration(seconds) * time.Second
	}
}

// WithRetries tells the Api client how many retries to perform when hit a rate limit
func WithRetry(tries int) Option {
	return func(c *Client) {
		c.tries = tries
	}
}

// Body is an API request/response body
type Body map[string]interface{}

// shopify API client
type Client struct {
	client         *http.Client
	domain         string
	baseUrl        string
	accessToken    string
	version        string
	tries          int
	availableLimit int // used for handling rate limits
}

// Create a new shopify Api client
// the domain parameter is the shop domain
func NewClient(domain, accessToken string, opts ...Option) *Client {
	baseUrl := fmt.Sprintf("https://%s/admin/api/%s", domain, defaultApiVersion)
	client := http.Client{
		Timeout: defaultTimeout,
	}
	c := &Client{
		client:         &client,
		domain:         domain,
		baseUrl:        baseUrl,
		accessToken:    accessToken,
		version:        defaultApiVersion,
		tries:          defaultRetries,
		availableLimit: 0,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) newRequest(method string, path string, body Body) (*http.Request, error) {
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

	return req, nil
}

func (c *Client) rest(method string, path string, body Body) (Body, error) {
	threshold := 2 // rate limit threshold
	req, err := c.newRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})

	for t := 1; t <= c.tries; t++ {
		res, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusTooManyRequests {
			retryAfter := res.Header.Get("Retry-After")
			if t == c.tries {
				return nil, ErrRateLimit
			}
			r, _ := strconv.Atoi(retryAfter)
			time.Sleep(time.Second * time.Duration(r))
			continue
		}

		limit := res.Header.Get("X-Shopify-Shop-Api-Call-Limit")
		s := strings.Split(limit, "/")
		bucketSize, _ := strconv.Atoi(s[1])
		requestCount, _ := strconv.Atoi(s[0])
		c.availableLimit = bucketSize - requestCount

		if c.availableLimit < threshold {
			time.Sleep(time.Second * 2)
			continue
		}

		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// checks if a graphql response has a rate limit error and return it if exists
func (c *Client) rateLimitError(errors []interface{}) bool {
	for _, err := range errors {
		if ext, ok := err.(map[string]interface{}); ok {
			if e, ok := ext["extensions"].(map[string]interface{}); ok {
				code, ok := e["code"].(string)
				if ok && (code == "MAX_COST_EXCEEDED" || code == "THROTTLED") {
					return true
				}
			}
		}
	}
	return false
}

// get the current available rate limit from the graphql response
func (c *Client) graphqlAvailableLimit(extentions map[string]interface{}) int {
	if cost, ok := extentions["cost"].(map[string]interface{}); ok {
		if throttleStatus, ok := cost["throttleStatus"].(map[string]interface{}); ok {
			if currentlyAvailable, ok := throttleStatus["currentlyAvailable"].(float64); ok {
				return int(currentlyAvailable)
			}
		}
	}
	return -1
}

func (c *Client) graphql(body Body) (Body, error) {
	threshold := 50
	req, err := c.newRequest(http.MethodPost, "graphql.json", body)
	if err != nil {
		return nil, err
	}

	for t := 1; t <= c.tries; t++ {
		res, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected server response: %s", res.Status)
		}
		result := make(map[string]interface{})
		dec := json.NewDecoder(res.Body)
		if err := dec.Decode(&result); err != nil {
			return nil, err
		}

		// handle errors
		if e, ok := result["errors"]; ok {
			errs := e.([]interface{})
			// rate limit error
			if c.rateLimitError(errs) {
				if t == c.tries {
					return nil, ErrRateLimit
				}
				time.Sleep(2 * time.Second)
				continue
			} else {
				if errMessage, ok := errs[0].(map[string]interface{})["message"].(string); ok {
					return nil, GraphqlError{msg: errMessage}
				}
			}
		}

		// check available rate limit
		if e, ok := result["extensions"]; ok {
			l := c.graphqlAvailableLimit(e.(map[string]interface{}))
			if l != -1 {
				c.availableLimit = l
			}
		}
		if c.availableLimit < threshold {
			time.Sleep(2 * time.Second)
			continue
		}
		return result["data"].(map[string]interface{}), nil
	}
	return nil, nil
}

// Get performs a get request and returns the result
func (c *Client) Get(path string) (Body, error) {
	return c.rest(http.MethodGet, path, nil)
}

// post performs a post request and returns the result
func (c *Client) Post(path string, body Body) (Body, error) {
	return c.rest(http.MethodPost, path, body)
}

// Put performs a put request and returns the result
func (c *Client) Put(path string, body Body) (Body, error) {
	return c.rest(http.MethodPut, path, body)
}

// Delete performs a delete request
func (c *Client) Delete(path string) error {
	_, err := c.rest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return nil
}

// Graphql sends a graphql query to shopify admin api.
func (c *Client) Graphql(query string, variables map[string]interface{}) (Body, error) {
	body := map[string]interface{}{
		"query": query,
	}
	if variables != nil {
		body["variables"] = variables
	}
	return c.graphql(body)
}
