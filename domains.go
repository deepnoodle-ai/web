package web

import (
	"net/url"
	"strings"
)

// AreSameHost checks if two URLs have the same host value.
func AreSameHost(url1, url2 *url.URL) bool {
	return url1 != nil && url2 != nil && url1.Host == url2.Host
}

// AreRelatedHosts checks if two URLs are the same or are related by a common
// parent domain.
func AreRelatedHosts(url1, url2 *url.URL) bool {
	if url1 == nil || url2 == nil {
		return false
	}
	parts1 := strings.Split(url1.Host, ".")
	parts2 := strings.Split(url2.Host, ".")

	// Get the base domain (last two parts)
	if len(parts1) < 2 || len(parts2) < 2 {
		return false
	}
	base1 := strings.Join(parts1[len(parts1)-2:], ".")
	base2 := strings.Join(parts2[len(parts2)-2:], ".")
	return base1 == base2
}
