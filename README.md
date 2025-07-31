# Web Utilities

A lightweight Go library for working with web pages, URLs, and HTML documents.

## Installation

```bash
go get github.com/deepnoodle-ai/web
```

## Quick Start

### Text Normalization

```go
package main

import (
    "fmt"
    "github.com/deepnoodle-ai/web"
)

func main() {
    // Clean up messy text
    text := "  Hello\t&nbsp;World!  \n\n  "
    clean := web.NormalizeText(text)
    fmt.Println(clean) // "Hello World!"
    
    // Normalize URLs
    url, _ := web.NormalizeURL("example.com/path/")
    fmt.Println(url.String()) // "https://example.com/path"
}
```

### HTML Document Parsing

```go
package main

import (
    "fmt"
    "github.com/deepnoodle-ai/web"
)

func main() {
    html := `
    <html>
        <head>
            <title>My Blog Post</title>
            <meta name="description" content="An amazing article">
            <meta property="og:image" content="https://example.com/image.jpg">
        </head>
        <body>
            <h1>Welcome</h1>
            <p>This is a great article about web scraping.</p>
            <a href="https://example.com">Visit our site</a>
        </body>
    </html>`
    
    doc, err := web.NewDocument(html)
    if err != nil {
        panic(err)
    }
    
    // Extract metadata
    fmt.Println("Title:", doc.Title())             // "My Blog Post"
    fmt.Println("Description:", doc.Description()) // "An amazing article"
    fmt.Println("Image:", doc.Image())             // "https://example.com/image.jpg"
    fmt.Println("H1:", doc.H1())                   // "Welcome"
    
    // Get all links
    links := doc.Links()
    for _, link := range links {
        fmt.Printf("Link: %s -> %s\n", link.Text, link.URL)
    }
    
    // Get clean paragraphs
    paragraphs := doc.Paragraphs()
    for _, p := range paragraphs {
        fmt.Println("Paragraph:", p)
    }
}
```

### URL and Domain Utilities

```go
package main

import (
    "fmt"
    "net/url"
    "github.com/deepnoodle-ai/web"
)

func main() {
    // Check if URL points to media
    u, _ := url.Parse("https://example.com/video.mp4")
    isMedia := web.IsMediaURL(u)
    fmt.Println("Is media:", isMedia) // true
    
    // Compare domains
    u1, _ := url.Parse("https://blog.example.com")
    u2, _ := url.Parse("https://shop.example.com") 
    related := web.AreRelatedHosts(u1, u2)
    fmt.Println("Related domains:", related) // true
    
    // Smart text chunking
    longText := "This is a very long article. It has multiple sentences. We want to split it intelligently."
    chunks := web.Chunk(longText, 30)
    for i, chunk := range chunks {
        fmt.Printf("Chunk %d: %s\n", i+1, chunk)
    }
}
```

## API Reference

### Text Processing

| Function | Description |
|----------|-------------|
| `NormalizeText(text string) string` | Clean and normalize text by removing extra whitespace and non-printable characters |
| `Chunk(text string, size int) []string` | Split text into chunks with intelligent boundary detection |
| `EndsWithPunctuation(s string) bool` | Check if a string ends with common punctuation |

### URL Utilities

| Function | Description |
|----------|-------------|
| `NormalizeURL(value string) (*url.URL, error)` | Parse and normalize a URL string |
| `IsMediaURL(u *url.URL) bool` | Check if URL points to a media file |
| `AreSameHost(url1, url2 *url.URL) bool` | Check if two URLs have the same host |
| `AreRelatedHosts(url1, url2 *url.URL) bool` | Check if two URLs share the same base domain |
| `SortURLs(urls []*url.URL)` | Sort URLs alphabetically |

### Document Extraction

| Method | Description |
|--------|-------------|
| `NewDocument(html string) (*Document, error)` | Create a new document parser |
| `Title() string` | Extract page title (with fallbacks) |
| `Description() string` | Get meta description |
| `Image() string` | Get Open Graph image URL |
| `Canonical() string` | Get canonical URL |
| `H1() string` | Get first H1 heading |
| `Lang() string` | Get document language |
| `Author() string` | Get author metadata |
| `Keywords() []string` | Extract keywords |
| `PublishedTime() time.Time` | Get publication date |
| `Links() []*Link` | Get all links with text |
| `Images() []*Link` | Get all images with alt text |
| `Paragraphs() []string` | Extract all paragraph text |
| `HTML(options HTMLOptions) (string, error)` | Get cleaned/prettified HTML |

## HTML Cleaning

Clean up HTML before processing with predefined removal patterns:

```go
options := web.HTMLOptions{
    RemoveElements: web.StandardRemoveElements, // Remove scripts, styles, modals, etc.
    Prettify:       true,                       // Format HTML nicely
}

cleanHTML, err := doc.HTML(options)
```

The `StandardRemoveElements` includes common unwanted elements like scripts,
styles, cookie banners, modals, and form inputs.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
