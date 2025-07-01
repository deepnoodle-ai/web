package web

import "github.com/yosssi/gohtml"

// FormatHTML parses the input HTML string, formats it and returns the result.
func FormatHTML(html string) string {
	return gohtml.Format(html)
}
