package web

import (
	"net/url"
	"os"
	"strings"
)

func ReadFileItems(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var items []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		items = append(items, line)
	}
	return items, nil
}

func ResolveLink(domain, value string) (string, bool) {
	// Parse the input URL
	parsedURL, err := url.Parse(value)
	if err != nil {
		return "", false
	}

	// Remove fragment
	parsedURL.Fragment = ""

	// Check if it's already absolute
	if parsedURL.IsAbs() {
		// Only accept HTTP/HTTPS schemes
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return "", false
		}
		// Normalize and return
		normalizedURL, err := NormalizeURL(parsedURL.String())
		if err != nil {
			return "", false
		}
		return normalizedURL.String(), true
	}

	// For relative URLs, we need to resolve against the domain
	// First, ensure domain has a scheme
	baseDomain := domain
	if !strings.HasPrefix(baseDomain, "http://") && !strings.HasPrefix(baseDomain, "https://") {
		baseDomain = "https://" + baseDomain
	}

	// Parse the base domain
	baseURL, err := url.Parse(baseDomain)
	if err != nil {
		return "", false
	}

	// Resolve the relative URL against the base
	resolvedURL := baseURL.ResolveReference(parsedURL)

	// Normalize and return
	normalizedURL, err := NormalizeURL(resolvedURL.String())
	if err != nil {
		return "", false
	}
	return normalizedURL.String(), true
}
