package web

import (
	"fmt"
	"html"
	"net/url"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

func removeNonPrintableChars(input string) string {
	var builder strings.Builder
	for _, r := range input {
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			builder.WriteRune(r)
		} else {
			builder.WriteRune(' ')
		}
	}
	return builder.String()
}

// NormalizeText applies transformations to the given text that are commonly
// helpful for cleaning up text read from a webpage.
// - Trim whitespace
// - Unescape HTML entities
// - Remove non-printable characters
func NormalizeText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return text
	}
	text = html.UnescapeString(text)
	text = removeNonPrintableChars(text)
	return text
}

// NormalizeURL parses a URL string and returns a normalized URL. The following
// transformations are applied:
// - Trim whitespace
// - Convert http:// to https://
// - Add https:// prefix if missing
// - Remove any query parameters and URL fragments
func NormalizeURL(value string) (*url.URL, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("invalid empty url")
	}
	if !strings.HasPrefix(value, "http") {
		if strings.Contains(value, "://") {
			return nil, fmt.Errorf("invalid url: %s", value)
		}
		value = "https://" + value
	}
	if strings.HasPrefix(value, "http://") {
		value = "https://" + value[7:]
	}
	u, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("invalid url %q: %w", value, err)
	}
	u.ForceQuery = false
	u.RawQuery = ""
	u.Fragment = ""
	if u.Path == "/" {
		u.Path = ""
	}
	return u, nil
}

// SortURLs sorts a slice of URLs by their string representation.
func SortURLs(urls []*url.URL) {
	sort.Slice(urls, func(i, j int) bool {
		return urls[i].String() < urls[j].String()
	})
}

var punctuation = map[rune]bool{
	'.':  true,
	',':  true,
	':':  true,
	';':  true,
	'?':  true,
	'!':  true,
	'"':  true,
	'\'': true,
}

// EndsWithPunctuation checks if a string ends with a punctuation mark.
func EndsWithPunctuation(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Get the last rune efficiently without converting the entire string
	lastRune, size := utf8.DecodeLastRuneInString(s)
	if size == 0 {
		return false
	}
	return punctuation[lastRune]
}
