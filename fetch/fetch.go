package fetch

import "context"

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
}

// Metadata conveys high level information about a page.
type Metadata struct {
	Title         string `json:"title,omitempty"`
	Description   string `json:"description,omitempty"`
	Language      string `json:"language,omitempty"`
	Keywords      string `json:"keywords,omitempty"`
	Author        string `json:"author,omitempty"`
	Canonical     string `json:"canonical,omitempty"`
	Heading       string `json:"heading,omitempty"`
	Robots        string `json:"robots,omitempty"`
	Image         string `json:"image,omitempty"`
	Icon          string `json:"icon,omitempty"`
	PublishedTime string `json:"published_time,omitempty"`
	Tags          []Meta `json:"tags,omitempty"`
}

// Link represents a link on a page.
type Link struct {
	URL  string `json:"url"`
	Text string `json:"text,omitempty"`
}

// Meta represents a meta tag on a page.
type Meta struct {
	Tag      string `json:"tag"`
	Name     string `json:"name,omitempty"`
	Content  string `json:"content,omitempty"`
	Charset  string `json:"charset,omitempty"`
	Property string `json:"property,omitempty"`
}

// Response defines the JSON payload for fetch responses.
type Response struct {
	URL        string            `json:"url"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	HTML       string            `json:"html,omitempty"`
	Markdown   string            `json:"markdown,omitempty"`
	Screenshot string            `json:"screenshot,omitempty"`
	PDF        string            `json:"pdf,omitempty"`
	Error      string            `json:"error,omitempty"`
	Metadata   Metadata          `json:"metadata,omitempty"`
	Links      []Link            `json:"links,omitempty"`
}

// Fetcher defines an interface for fetching pages.
type Fetcher interface {
	Fetch(ctx context.Context, request *Request) (*Response, error)
}
