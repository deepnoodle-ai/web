package fetch

import (
	"fmt"
	"strings"

	"github.com/myzie/web"
)

// ProcessRequest applies request options to the given HTML content and builds
// the corresponding response. Applies any requested transformations. This is
// a reference implementation and may not be used in all cases.
func ProcessRequest(request *Request, html string) (*Response, error) {
	html = strings.TrimSpace(html)
	if html == "" {
		return &Response{
			URL:        request.URL,
			StatusCode: 200,
		}, nil
	}

	// Parse the HTML
	doc, err := web.NewDocument(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}
	metadata := doc.Metadata()

	// Render transformed HTML with options
	renderedHTML, err := doc.Render(web.RenderOptions{
		Prettify:    request.Prettify,
		ExcludeTags: request.ExcludeTags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render html: %w", err)
	}

	// By default, return the HTML but not markdown
	includeHTML := true
	includeMarkdown := false

	// Specified formats were requested
	if len(request.Formats) > 0 {
		includeHTML = false
		for _, format := range request.Formats {
			switch format {
			case "markdown":
				includeMarkdown = true
			case "html":
				includeHTML = true
			}
		}
	}

	// Generate markdown if requested
	var markdownContent string
	if includeMarkdown {
		markdownContent, err = web.Markdown(renderedHTML)
		if err != nil {
			return nil, fmt.Errorf("failed to generate markdown: %w", err)
		}
	}

	// Decide whether to include the HTML
	if !includeHTML {
		renderedHTML = ""
	}

	// Massage link types
	var links []*Link
	for _, link := range doc.Links() {
		links = append(links, &Link{URL: link.URL, Text: link.Text})
	}

	return &Response{
		URL:        request.URL,
		StatusCode: 200,
		Headers:    map[string]string{},
		HTML:       renderedHTML,
		Markdown:   markdownContent,
		Metadata:   Metadata(metadata),
		Links:      links,
	}, nil
}
