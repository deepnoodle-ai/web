package web

import (
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/yosssi/gohtml"
)

// FormatHTML parses the input HTML string, formats it and returns the result.
func FormatHTML(html string) string {
	return gohtml.Format(html)
}

// Markdown converts HTML to Markdown.
func Markdown(html string) (string, error) {
	return htmltomarkdown.ConvertString(html)
}
