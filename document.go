package web

import (
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Link represents a link on a page.
type Link struct {
	URL  string `json:"url"`
	Text string `json:"text,omitempty"`
}

// Host returns the host of the link.
func (l *Link) Host() string {
	u, err := url.Parse(l.URL)
	if err != nil {
		return ""
	}
	return u.Host
}

// Meta represents a meta tag on a page.
type Meta struct {
	Tag      string `json:"tag"`
	Name     string `json:"name,omitempty"`
	Property string `json:"property,omitempty"`
	Content  string `json:"content,omitempty"`
	Charset  string `json:"charset,omitempty"`
}

// Document helps parse and extract information from an HTML document.
type Document struct {
	doc  *goquery.Document
	text string
}

// NewDocument creates a new Document from an HTML string.
func NewDocument(text string) (*Document, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(text))
	if err != nil {
		return nil, err
	}
	return &Document{doc: doc, text: text}, nil
}

// Raw returns the raw HTML text of the document.
func (d *Document) Raw() string {
	return d.text
}

// GoqueryDocument returns the underlying goquery document.
func (d *Document) GoqueryDocument() *goquery.Document {
	return d.doc
}

// Lang returns the language of the document.
func (d *Document) Lang() string {
	if s := d.doc.Find("html").First(); len(s.Nodes) > 0 {
		return strings.ToLower(strings.TrimSpace(s.AttrOr("lang", "")))
	}
	return ""
}

// Canonical returns the canonical URL of the document.
func (d *Document) Canonical() string {
	if s := d.doc.Find("link[rel='canonical']"); len(s.Nodes) > 0 {
		return strings.TrimSpace(s.AttrOr("href", ""))
	}
	return ""
}

// Title returns the title of the document.
func (d *Document) Title() string {
	if s := d.doc.Find("title").First(); len(s.Nodes) > 0 {
		return NormalizeText(s.Text())
	}
	if s := d.doc.Find("meta[property='og:title']").First(); len(s.Nodes) > 0 {
		return NormalizeText(s.AttrOr("content", ""))
	}
	if s := d.doc.Find("meta[name='title']").First(); len(s.Nodes) > 0 {
		return NormalizeText(s.AttrOr("content", ""))
	}
	return ""
}

// H1 returns the first H1 element of the document.
func (d *Document) H1() string {
	var h1 string
	d.doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		h1 = NormalizeText(s.Text())
	})
	return h1
}

// Robots returns the robots meta tag of the document.
func (d *Document) Robots() string {
	if s := d.doc.Find("meta[name='robots']").First(); len(s.Nodes) > 0 {
		return strings.TrimSpace(s.AttrOr("content", ""))
	}
	return ""
}

// Description returns the description meta tag of the document.
func (d *Document) Description() string {
	if s := d.doc.Find("meta[name='description']"); len(s.Nodes) > 0 {
		return NormalizeText(s.AttrOr("content", ""))
	}
	if s := d.doc.Find("meta[property='og:description']"); len(s.Nodes) > 0 {
		return NormalizeText(s.AttrOr("content", ""))
	}
	return ""
}

// Image returns the image meta tag of the document.
func (d *Document) Image() string {
	if s := d.doc.Find("meta[property='og:image']").First(); len(s.Nodes) > 0 {
		return strings.TrimSpace(s.AttrOr("content", ""))
	}
	if s := d.doc.Find("meta[property='og:image:url']").First(); len(s.Nodes) > 0 {
		return strings.TrimSpace(s.AttrOr("content", ""))
	}
	return ""
}

// Icon returns the icon link of the document.
func (d *Document) Icon() string {
	if s := d.doc.Find("link[rel='icon']").First(); len(s.Nodes) > 0 {
		return strings.TrimSpace(s.AttrOr("href", ""))
	}
	if s := d.doc.Find("link[rel='shortcut icon']").First(); len(s.Nodes) > 0 {
		return strings.TrimSpace(s.AttrOr("href", ""))
	}
	return ""
}

// Keywords returns the keywords meta tag of the document.
func (d *Document) Keywords() []string {
	if s := d.doc.Find("meta[name='keywords']").First(); len(s.Nodes) > 0 {
		keywords := s.AttrOr("content", "")
		if len(keywords) > 0 {
			return parseKeywords(keywords)
		}
	}
	if s := d.doc.Find("meta[property='og:keywords']").First(); len(s.Nodes) > 0 {
		keywords := s.AttrOr("content", "")
		return parseKeywords(keywords)
	}
	return []string{}
}

// Author returns the author meta tag of the document.
func (d *Document) Author() string {
	if s := d.doc.Find("meta[name='author']").First(); len(s.Nodes) > 0 {
		return strings.TrimSpace(s.AttrOr("content", ""))
	}
	if s := d.doc.Find("meta[property='og:author']").First(); len(s.Nodes) > 0 {
		return strings.TrimSpace(s.AttrOr("content", ""))
	}
	return ""
}

// TwitterSite returns the twitter site meta tag of the document.
func (d *Document) TwitterSite() string {
	if s := d.doc.Find("meta[name='twitter:site']").First(); len(s.Nodes) > 0 {
		return strings.TrimSpace(s.AttrOr("content", ""))
	}
	if s := d.doc.Find("meta[property='twitter:site']").First(); len(s.Nodes) > 0 {
		return strings.TrimSpace(s.AttrOr("content", ""))
	}
	return ""
}

// PublishedTime returns the published time meta tag of the document.
func (d *Document) PublishedTime() time.Time {
	var timeStr string
	d.doc.Find("meta[name='article:published_time']").Each(func(i int, s *goquery.Selection) {
		timeStr = strings.TrimSpace(s.AttrOr("content", ""))
	})
	if timeStr != "" {
		value, _ := time.Parse(time.RFC3339, timeStr)
		return value
	}
	d.doc.Find("meta[property='article:published_time']").Each(func(i int, s *goquery.Selection) {
		timeStr = strings.TrimSpace(s.AttrOr("content", ""))
	})
	if timeStr != "" {
		value, _ := time.Parse(time.RFC3339, timeStr)
		return value
	}
	d.doc.Find("meta[property='og:published_time']").Each(func(i int, s *goquery.Selection) {
		timeStr = strings.TrimSpace(s.AttrOr("content", ""))
	})
	value, _ := time.Parse(time.RFC3339, timeStr)
	return value
}

// Meta returns the meta tags of the document.
func (d *Document) Meta() []*Meta {
	metas := []*Meta{}
	d.doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		var meta Meta
		meta.Tag = "meta"
		meta.Name = s.AttrOr("name", "")
		meta.Property = s.AttrOr("property", "")
		meta.Content = s.AttrOr("content", "")
		meta.Charset = s.AttrOr("charset", "")
		metas = append(metas, &meta)
	})
	return metas
}

// Links returns the links on the document.
func (d *Document) Links() []*Link {
	links := []*Link{}
	d.doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		if href == "" {
			return
		}
		links = append(links, &Link{URL: href, Text: s.Text()})
	})
	return links
}

// Images returns the images on the document.
func (d *Document) Images() []*Link {
	images := []*Link{}
	d.doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src := s.AttrOr("src", "")
		if src == "" {
			return
		}
		images = append(images, &Link{URL: src, Text: s.AttrOr("alt", "")})
	})
	return images
}

// Paragraphs returns the paragraphs on the document.
func (d *Document) Paragraphs() []string {
	paragraphs := []string{}
	d.doc.Find("p").Each(func(i int, s *goquery.Selection) {
		nodeText := strings.TrimSpace(s.Text())
		if nodeText == "" {
			return
		}
		paragraphs = append(paragraphs, nodeText)
	})
	return paragraphs
}

// HTML returns the HTML of the document, with optional filtering and prettifying.
func (d *Document) HTML(options HTMLOptions) (string, error) {
	// Return the raw text if no options are provided
	if len(options.RemoveElements) == 0 && !options.Prettify {
		return d.text, nil
	}

	text := d.text

	// Optionally remove elements
	if options.RemoveElements != nil {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(d.text))
		if err != nil {
			return "", err
		}
		for _, selector := range options.RemoveElements {
			doc.Find(selector).Remove()
		}
		text, err = doc.Html()
		if err != nil {
			return "", err
		}
	}

	// Optionally prettify
	if options.Prettify {
		text = FormatHTML(text)
	}

	return text, nil
}

// HTMLOptions contains options for generating markdown from a document.
type HTMLOptions struct {
	RemoveElements []string
	Prettify       bool
}

// StandardRemoveElements contains the suggested elements to remove from HTML
// in order to clean it up before conversion to markdown.
var StandardRemoveElements = []string{
	`[role="dialog"]`,
	`[aria-modal="true"]`,
	`[id*="cookie"]`,
	`[id*="popup"]`,
	`[id*="modal"]`,
	`[class*="modal"]`,
	`[class*="dialog"]`,
	"img[data-cookieconsent]",
	"script",
	"style",
	"hr",
	"noscript",
	"iframe",
	"select",
	"input",
	"button",
	"svg",
	"form",
}

// parseKeywords parses the keywords from a string.
func parseKeywords(s string) []string {
	if s == "" {
		return []string{}
	}
	s = strings.ToLower(s)
	splitChar := " "
	if strings.Contains(s, ",") {
		splitChar = ","
	}
	kws := map[string]bool{}
	for _, kw := range strings.Split(s, splitChar) {
		if value := strings.TrimSpace(kw); value != "" {
			kws[value] = true
		}
	}
	results := []string{}
	for kw := range kws {
		results = append(results, kw)
	}
	sort.Strings(results)
	return results
}
