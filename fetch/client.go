package fetch

import (
	"bytes"
	"context"
	"encoding/json"

	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/myzie/web/errors"
)

// ClientOptions defines the options for the client.
type ClientOptions struct {
	BaseURL   string            // Optional proxy base URL
	AuthToken string            // Optional authorization token
	Timeout   time.Duration     // Optional HTTP timeout
	Headers   map[string]string // Optional HTTP headers
}

// Client defines a client for fetching pages via a remote proxy.
type Client struct {
	baseURL    string
	authToken  string
	httpClient *http.Client
	headers    map[string]string
}

// NewClient creates a new client with the given options.
func NewClient(options ClientOptions) *Client {
	timeout := options.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second // Default 30 second timeout
	}
	return &Client{
		baseURL:   options.BaseURL,
		authToken: options.AuthToken,
		headers:   options.Headers,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SetHeader sets a header for the client.
func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

// Fetch a page using a remote proxy.
func (c *Client) Fetch(ctx context.Context, request *Request) (*Response, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	for key, value := range c.headers {
		httpReq.Header.Set(key, value)
	}
	if c.authToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer httpResp.Body.Close()

	responseBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if httpResp.StatusCode != 200 {
		err := fmt.Errorf("request failed with status %d: %s",
			httpResp.StatusCode, string(responseBody))
		return nil, errors.NewRequestError(err).
			WithStatusCode(httpResp.StatusCode).
			WithRawURL(c.baseURL)
	}

	var response Response
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &response, nil
}
