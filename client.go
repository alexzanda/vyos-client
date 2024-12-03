package vyos

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client provides client Methods
type Client struct {
	client *http.Client
	Cfg    Config
}

// Config provides configuration for the client. These values are only read in
// when NewClient is called.
type Config struct {
	Host    string
	ApiKey  string
	SkipTLS bool
	Timeout time.Duration
}

// NewClient constructs a new Client
func NewClient(config Config) *Client {
	httpclient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: config.SkipTLS},
		},
	}

	newClient := &Client{
		Cfg:    config,
		client: httpclient,
	}
	return newClient
}

func (c *Client) doRequest(method, endpoint string, queryMap map[string]string, body []byte, contentType string) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", c.Cfg.Host, endpoint), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for key, value := range queryMap {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	if len(contentType) != 0 {
		req.Header.Add("Content-Type", contentType)
	}

	return c.client.Do(req)
}

func (c *Client) post(endpoint string, queryMap map[string]string, body []byte, contentType string) ([]byte, error) {
	res, err := c.doRequest(http.MethodPost, endpoint, queryMap, body, contentType)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()
	respBody, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return respBody, fmt.Errorf("non 2xx response code received, code: %d, resp: %s", res.StatusCode, string(respBody))
	}

	resp := new(apiResponse)
	if err = json.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}

	return respBody, nil
}

type apiResponse struct {
	Success bool `json:"success"`
	Error   any  `json:"error"`
	Data    any  `json:"data"`
}
