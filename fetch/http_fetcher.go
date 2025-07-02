package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultMaxBodySize = 10 * 1024 * 1024 // 10 MB
	DefaultTimeout     = 30 * time.Second
)

var (
	DefaultHTTPClient = &http.Client{Timeout: DefaultTimeout}
	DefaultHeaders    = map[string]string{}
)

// HTTPFetcherOptions defines the options for the HTTP fetcher.
type HTTPFetcherOptions struct {
	Timeout     time.Duration
	Headers     map[string]string
	Client      *http.Client
	MaxBodySize int64
}

// HTTPFetcher implements the Fetcher interface using standard HTTP client.
type HTTPFetcher struct {
	timeout     time.Duration
	headers     map[string]string
	client      *http.Client
	maxBodySize int64
}

// NewHTTPFetcher creates a new HTTP fetcher
func NewHTTPFetcher(options HTTPFetcherOptions) *HTTPFetcher {
	if options.Timeout == 0 {
		options.Timeout = DefaultTimeout
	}
	if options.Headers == nil {
		options.Headers = DefaultHeaders
	}
	if options.Client == nil {
		options.Client = DefaultHTTPClient
	}
	if options.MaxBodySize == 0 {
		options.MaxBodySize = DefaultMaxBodySize
	}
	return &HTTPFetcher{
		timeout:     options.Timeout,
		headers:     options.Headers,
		client:      options.Client,
		maxBodySize: options.MaxBodySize,
	}
}

// Fetch implements the Fetcher interface for HTTP requests
func (f *HTTPFetcher) Fetch(ctx context.Context, req *Request) (*Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, req.URL, nil)
	if err != nil {
		return nil, err
	}

	// Apply default headers
	for key, value := range f.headers {
		if httpReq.Header.Get(key) == "" {
			httpReq.Header.Set(key, value)
		}
	}

	// Apply custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := f.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Confirm the content type indicates HTML
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}

	// Use LimitReader to prevent reading excessive data
	limitedReader := io.LimitReader(resp.Body, f.maxBodySize+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	// Check if the body is too large
	if len(body) > int(f.maxBodySize) {
		return nil, fmt.Errorf("response size exceeds limit of %d bytes", f.maxBodySize)
	}

	// Convert response headers to map[string]string
	headers := make(map[string]string)
	for name, values := range resp.Header {
		if len(values) > 0 {
			headers[name] = values[0] // Use first value if multiple
		}
	}

	// Apply processing options
	response, err := ProcessRequest(req, string(body))
	if err != nil {
		return nil, err
	}

	// Set other response fields
	response.URL = req.URL
	response.StatusCode = resp.StatusCode
	response.Headers = headers
	return response, nil
}
