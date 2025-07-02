package web

// Metadata conveys high level information about a page.
type Metadata struct {
	Title         string   `json:"title,omitempty"`
	Description   string   `json:"description,omitempty"`
	Language      string   `json:"language,omitempty"`
	Author        string   `json:"author,omitempty"`
	CanonicalURL  string   `json:"canonical_url,omitempty"`
	Heading       string   `json:"heading,omitempty"`
	Robots        string   `json:"robots,omitempty"`
	Image         string   `json:"image,omitempty"`
	Icon          string   `json:"icon,omitempty"`
	PublishedTime string   `json:"published_time,omitempty"`
	Keywords      []string `json:"keywords,omitempty"`
	Tags          []*Meta  `json:"tags,omitempty"`
}
