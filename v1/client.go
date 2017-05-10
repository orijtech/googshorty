package googshorty

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/orijtech/otils"
)

type Client struct {
	sync.RWMutex
	_apiKey string

	rt http.RoundTripper
}

func (c *Client) apiKey() string {
	c.RLock()
	defer c.RUnlock()
	return c._apiKey
}

func (c *Client) SetHTTPRoundTripper(rt http.RoundTripper) {
	c.Lock()
	defer c.Unlock()

	c.rt = rt
}

func (c *Client) httpClient() *http.Client {
	c.RLock()
	rt := c.rt
	c.RUnlock()

	if rt == nil {
		rt = http.DefaultTransport
	}

	return &http.Client{Transport: rt}
}

const (
	envAPIClientKeyKey = "GOOGLE_URL_SHORTENER_API_KEY"
)

func NewClient(apiKeys ...string) (*Client, error) {
	apiKey := otils.FirstNonEmptyString(apiKeys...)
	if apiKey != "" {
		return &Client{_apiKey: apiKey}, nil
	}

	// Otherwise fallback to retrieving the key from the environment
	return NewClientFromEnv()
}

var errUnsetEnvClientKey = fmt.Errorf("could not find %q in your environment", envAPIClientKeyKey)

func NewClientFromEnv() (*Client, error) {
	apiKey := os.Getenv(envAPIClientKeyKey)
	if apiKey == "" {
		return nil, errUnsetEnvClientKey
	}
	return &Client{_apiKey: apiKey}, nil
}
