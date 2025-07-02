package fetch

// FakeUserAgent may be used to mimic a real browser.
const FakeUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0"

// FakeHeaders may be used to mimic a real browser.
var FakeHeaders = map[string]string{
	"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
	"Accept-Language":           "en-US,en;q=0.5",
	"Connection":                "keep-alive",
	"Dnt":                       "1",
	"Sec-Fetch-Dest":            "document",
	"Sec-Fetch-Mode":            "navigate",
	"Sec-Fetch-Site":            "cross-site",
	"Upgrade-Insecure-Requests": "1",
	"User-Agent":                FakeUserAgent,
	"Priority":                  "u=0, i",
}
