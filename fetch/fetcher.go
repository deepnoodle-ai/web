package fetch

import (
	"context"
	"time"

	"github.com/myzie/web"
)

// Type aliases for convenience.
type (
	Link     web.Link
	Meta     web.Meta
	Metadata web.Metadata
)

// Request defines the JSON payload for fetch requests.
type Request struct {
	URL             string            `json:"url"`
	OnlyMainContent bool              `json:"only_main_content,omitempty"`
	IncludeTags     []string          `json:"include_tags,omitempty"`
	ExcludeTags     []string          `json:"exclude_tags,omitempty"`
	MaxAge          int               `json:"max_age,omitempty"`  // milliseconds
	Timeout         int               `json:"timeout,omitempty"`  // milliseconds
	WaitFor         int               `json:"wait_for,omitempty"` // milliseconds
	Fetcher         string            `json:"fetcher,omitempty"`
	Mobile          bool              `json:"mobile,omitempty"`
	Prettify        bool              `json:"prettify,omitempty"`
	Formats         []string          `json:"formats,omitempty"`
	Actions         []Action          `json:"actions,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	StorageState    map[string]any    `json:"storage_state,omitempty"`
}

// Response defines the JSON payload for fetch responses.
type Response struct {
	URL          string            `json:"url"`
	StatusCode   int               `json:"status_code"`
	Headers      map[string]string `json:"headers"`
	HTML         string            `json:"html,omitempty"`
	Markdown     string            `json:"markdown,omitempty"`
	Screenshot   string            `json:"screenshot,omitempty"`
	PDF          string            `json:"pdf,omitempty"`
	Error        string            `json:"error,omitempty"`
	Metadata     Metadata          `json:"metadata,omitempty"`
	Links        []*Link           `json:"links,omitempty"`
	StorageState map[string]any    `json:"storage_state,omitempty"`
	Timestamp    time.Time         `json:"timestamp,omitzero"`
}

// Fetcher defines an interface for fetching pages.
type Fetcher interface {

	// Fetch a webpage and return the response.
	Fetch(ctx context.Context, request *Request) (*Response, error)
}
