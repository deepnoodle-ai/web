package web

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveNonPrintableChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "text with tabs and newlines",
			input:    "Hello\tWorld\n",
			expected: "Hello\tWorld\n",
		},
		{
			name:     "text with non-printable chars",
			input:    "Hello\x00\x01World",
			expected: "Hello  World",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only non-printable chars",
			input:    "\x00\x01\x02",
			expected: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeNonPrintableChars(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "text needing trimming",
			input:    "  Hello World  ",
			expected: "Hello World",
		},
		{
			name:     "text with HTML entities",
			input:    "Hello &amp; World &lt;test&gt;",
			expected: "Hello & World <test>",
		},
		{
			name:     "text with special quotes",
			input:    `"Hello" 'World'`,
			expected: "\"Hello\" 'World'",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \t\n  ",
			expected: "",
		},
		{
			name:     "text with non-printable chars",
			input:    "Hello\x00World",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeText(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "simple https URL",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "http URL converted to https",
			input:    "http://example.com",
			expected: "https://example.com",
		},
		{
			name:     "URL without protocol",
			input:    "example.com",
			expected: "https://example.com",
		},
		{
			name:     "URL with path",
			input:    "https://example.com/path",
			expected: "https://example.com/path",
		},
		{
			name:     "URL with root path removed",
			input:    "https://example.com/",
			expected: "https://example.com",
		},
		{
			name:     "URL with query and fragment removed",
			input:    "https://example.com/path?query=1#fragment",
			expected: "https://example.com/path",
		},
		{
			name:     "URL with whitespace",
			input:    "  https://example.com  ",
			expected: "https://example.com",
		},
		{
			name:        "empty URL",
			input:       "",
			expectError: true,
		},
		{
			name:        "invalid protocol",
			input:       "ftp://example.com",
			expectError: true,
		},
		{
			name:        "malformed URL",
			input:       "ht tp://example.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeURL(tt.input)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result.String())
			}
		})
	}
}

func TestAreSameHost(t *testing.T) {
	tests := []struct {
		name     string
		url1     string
		url2     string
		expected bool
	}{
		{
			name:     "same domain",
			url1:     "https://example.com/path1",
			url2:     "https://example.com/path2",
			expected: true,
		},
		{
			name:     "different domains",
			url1:     "https://example.com",
			url2:     "https://google.com",
			expected: false,
		},
		{
			name:     "same domain different subdomains",
			url1:     "https://www.example.com",
			url2:     "https://api.example.com",
			expected: false,
		},
		{
			name:     "nil URLs",
			url1:     "",
			url2:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u1, u2 *url.URL
			if tt.url1 != "" {
				u1, _ = url.Parse(tt.url1)
			}
			if tt.url2 != "" {
				u2, _ = url.Parse(tt.url2)
			}
			result := AreSameHost(u1, u2)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestAreRelatedHosts(t *testing.T) {
	tests := []struct {
		name     string
		url1     string
		url2     string
		expected bool
	}{
		{
			name:     "same domain",
			url1:     "https://example.com",
			url2:     "https://example.com",
			expected: true,
		},
		{
			name:     "related subdomains",
			url1:     "https://www.example.com",
			url2:     "https://api.example.com",
			expected: true,
		},
		{
			name:     "different base domains",
			url1:     "https://example.com",
			url2:     "https://google.com",
			expected: false,
		},
		{
			name:     "one URL is nil",
			url1:     "https://example.com",
			url2:     "",
			expected: false,
		},
		{
			name:     "both URLs are nil",
			url1:     "",
			url2:     "",
			expected: false,
		},
		{
			name:     "single part domains",
			url1:     "https://localhost",
			url2:     "https://localhost",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u1, u2 *url.URL
			if tt.url1 != "" {
				u1, _ = url.Parse(tt.url1)
			}
			if tt.url2 != "" {
				u2, _ = url.Parse(tt.url2)
			}
			result := AreRelatedHosts(u1, u2)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSortURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "sort URLs alphabetically",
			input:    []string{"https://z.com", "https://a.com", "https://m.com"},
			expected: []string{"https://a.com", "https://m.com", "https://z.com"},
		},
		{
			name:     "already sorted",
			input:    []string{"https://a.com", "https://b.com", "https://c.com"},
			expected: []string{"https://a.com", "https://b.com", "https://c.com"},
		},
		{
			name:     "single URL",
			input:    []string{"https://example.com"},
			expected: []string{"https://example.com"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert strings to URLs
			urls := make([]*url.URL, len(tt.input))
			for i, u := range tt.input {
				urls[i], _ = url.Parse(u)
			}

			// Sort the URLs
			SortURLs(urls)

			// Convert back to strings for comparison
			result := make([]string, len(urls))
			for i, u := range urls {
				result[i] = u.String()
			}

			require.Equal(t, tt.expected, result)
		})
	}
}

func TestEndsWithPunctuation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "ends with period",
			input:    "Hello.",
			expected: true,
		},
		{
			name:     "ends with comma",
			input:    "Hello,",
			expected: true,
		},
		{
			name:     "ends with question mark",
			input:    "Hello?",
			expected: true,
		},
		{
			name:     "ends with exclamation",
			input:    "Hello!",
			expected: true,
		},
		{
			name:     "ends with quote",
			input:    "Hello\"",
			expected: true,
		},
		{
			name:     "ends with apostrophe",
			input:    "Hello'",
			expected: true,
		},
		{
			name:     "ends with letter",
			input:    "Hello",
			expected: false,
		},
		{
			name:     "ends with number",
			input:    "Hello123",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "single punctuation",
			input:    ".",
			expected: true,
		},
		{
			name:     "unicode characters",
			input:    "Hello世界.",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndsWithPunctuation(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
