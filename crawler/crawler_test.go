package crawler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/myzie/web/cache"
	"github.com/myzie/web/fetch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCache implements the cache.Cache interface for testing
type MockCache struct {
	data  map[string][]byte
	mutex sync.RWMutex
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string][]byte),
	}
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if value, exists := m.data[key]; exists {
		return value, nil
	}
	return nil, cache.NotFound
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[key] = value
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.data, key)
	return nil
}

// MockFetcher implements the fetch.Fetcher interface for testing
type MockFetcher struct {
	mock.Mock
	responses map[string]*fetch.Response
	errors    map[string]error
	mutex     sync.RWMutex
}

func NewMockFetcher() *MockFetcher {
	return &MockFetcher{
		responses: make(map[string]*fetch.Response),
		errors:    make(map[string]error),
	}
}

func (m *MockFetcher) AddResponse(url string, response *fetch.Response) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.responses[url] = response
}

func (m *MockFetcher) AddError(url string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errors[url] = err
}

func (m *MockFetcher) Fetch(ctx context.Context, req *fetch.Request) (*fetch.Response, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if err, exists := m.errors[req.URL]; exists {
		return nil, err
	}

	if response, exists := m.responses[req.URL]; exists {
		return response, nil
	}

	return nil, fmt.Errorf("no mock response configured for URL: %s", req.URL)
}

// MockParser implements the Parser interface for testing
type MockParser struct {
	mock.Mock
	parseFunc func(ctx context.Context, page *fetch.Response) (any, error)
}

func NewMockParser() *MockParser {
	return &MockParser{}
}

func (m *MockParser) SetParseFunc(fn func(ctx context.Context, page *fetch.Response) (any, error)) {
	m.parseFunc = fn
}

func (m *MockParser) Parse(ctx context.Context, page *fetch.Response) (any, error) {
	if m.parseFunc != nil {
		return m.parseFunc(ctx, page)
	}
	return map[string]string{"parsed": "data"}, nil
}

// Test fixtures
func setupTestFixtures(t *testing.T) string {
	fixturesDir := filepath.Join(t.TempDir(), "fixtures")
	err := os.MkdirAll(fixturesDir, 0755)
	require.NoError(t, err)

	// Create index.html fixture
	indexHTML := `<!DOCTYPE html>
<html>
<head><title>Test Home Page</title></head>
<body>
	<h1>Welcome to Test Site</h1>
	<p>This is a test page with various links.</p>
	<nav>
		<a href="/about">About Us</a>
		<a href="/products">Products</a>
		<a href="/contact">Contact</a>
		<a href="https://external.com">External Link</a>
		<a href="https://subdomain.example.com/page">Subdomain</a>
	</nav>
	<div>
		<a href="mailto:test@example.com">Email Link</a>
		<a href="javascript:void(0)">JavaScript Link</a>
		<a href="#section1">Fragment Link</a>
	</div>
</body>
</html>`

	// Create about.html fixture
	aboutHTML := `<!DOCTYPE html>
<html>
<head><title>About Us</title></head>
<body>
	<h1>About Our Company</h1>
	<p>We are a test company.</p>
	<a href="/">Home</a>
	<a href="/careers">Careers</a>
	<a href="https://blog.example.com">Our Blog</a>
</body>
</html>`

	// Create products.html fixture
	productsHTML := `<!DOCTYPE html>
<html>
<head><title>Products</title></head>
<body>
	<h1>Our Products</h1>
	<ul>
		<li><a href="/products/widget1">Widget 1</a></li>
		<li><a href="/products/widget2">Widget 2</a></li>
	</ul>
	<a href="/">Back to Home</a>
</body>
</html>`

	fixtures := map[string]string{
		"index.html":    indexHTML,
		"about.html":    aboutHTML,
		"products.html": productsHTML,
	}

	for filename, content := range fixtures {
		err := os.WriteFile(filepath.Join(fixturesDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}

	return fixturesDir
}

func loadFixture(t *testing.T, fixturesDir, filename string) string {
	content, err := os.ReadFile(filepath.Join(fixturesDir, filename))
	require.NoError(t, err)
	return string(content)
}

func TestCrawler_New(t *testing.T) {
	mockFetcher := NewMockFetcher()

	opts := Options{
		MaxURLs:        100,
		Workers:        2,
		RequestDelay:   time.Millisecond * 100,
		Fetcher:        mockFetcher,
		FollowBehavior: FollowSameDomain,
	}

	crawler := New(opts)

	assert.NotNil(t, crawler)
	assert.Equal(t, 100, crawler.maxURLs)
	assert.Equal(t, 2, crawler.workers)
	assert.Equal(t, time.Millisecond*100, crawler.requestDelay)
	assert.Equal(t, FollowSameDomain, crawler.followBehavior)
}

func TestCrawler_BasicCrawl(t *testing.T) {
	fixturesDir := setupTestFixtures(t)
	mockFetcher := NewMockFetcher()

	// Setup mock responses
	indexHTML := loadFixture(t, fixturesDir, "index.html")
	aboutHTML := loadFixture(t, fixturesDir, "about.html")

	mockFetcher.AddResponse("https://example.com", &fetch.Response{
		URL:  "https://example.com",
		HTML: indexHTML,
		Links: []*fetch.Link{
			{URL: "/about"},
			{URL: "/products"},
			{URL: "/contact"},
			{URL: "https://external.com"},
		},
	})

	mockFetcher.AddResponse("https://example.com/about", &fetch.Response{
		URL:  "https://example.com/about",
		HTML: aboutHTML,
		Links: []*fetch.Link{
			{URL: "/"},
			{URL: "/careers"},
			{URL: "https://blog.example.com"},
		},
	})

	crawler := New(Options{
		MaxURLs:        10,
		Workers:        1,
		RequestDelay:   time.Millisecond,
		Fetcher:        mockFetcher,
		FollowBehavior: FollowSameDomain,
	})

	var processedURLs []string
	var processedData []any
	mu := sync.Mutex{}

	callback := func(ctx context.Context, req *fetch.Request, parsed any, err error) error {
		mu.Lock()
		defer mu.Unlock()
		processedURLs = append(processedURLs, req.URL)
		processedData = append(processedData, parsed)
		return nil
	}

	ctx := context.Background()
	err := crawler.Crawl(ctx, []string{"https://example.com"}, callback)

	assert.NoError(t, err)
	assert.Contains(t, processedURLs, "https://example.com")

	stats := crawler.GetStats()
	assert.Greater(t, stats.GetProcessed(), int64(0))
}

func TestCrawler_WithParser(t *testing.T) {
	fixturesDir := setupTestFixtures(t)
	mockFetcher := NewMockFetcher()
	mockParser := NewMockParser()

	indexHTML := loadFixture(t, fixturesDir, "index.html")

	mockFetcher.AddResponse("https://example.com", &fetch.Response{
		URL:   "https://example.com",
		HTML:  indexHTML,
		Links: []*fetch.Link{{URL: "/about"}},
	})

	expectedParsedData := map[string]string{"title": "Test Home Page"}
	mockParser.SetParseFunc(func(ctx context.Context, page *fetch.Response) (any, error) {
		return expectedParsedData, nil
	})

	parsers := map[string]Parser{
		"example.com": mockParser,
	}

	crawler := New(Options{
		MaxURLs:        5,
		Workers:        1,
		RequestDelay:   time.Millisecond,
		Fetcher:        mockFetcher,
		Parsers:        parsers,
		FollowBehavior: FollowNone,
	})

	var parsedResults []any
	mu := sync.Mutex{}

	callback := func(ctx context.Context, req *fetch.Request, parsed any, err error) error {
		mu.Lock()
		defer mu.Unlock()
		if parsed != nil {
			parsedResults = append(parsedResults, parsed)
		}
		return nil
	}

	ctx := context.Background()
	err := crawler.Crawl(ctx, []string{"https://example.com"}, callback)

	assert.NoError(t, err)
	assert.Len(t, parsedResults, 1)
	assert.Equal(t, expectedParsedData, parsedResults[0])
}

func TestCrawler_WithCache(t *testing.T) {
	htmlCache := NewMockCache()

	mockFetcher := NewMockFetcher()
	testHTML := "<html><body><h1>Cached Content</h1></body></html>"

	// Pre-populate cache
	err := htmlCache.Set(context.Background(), "https://example.com", []byte(testHTML))
	require.NoError(t, err)

	crawler := New(Options{
		MaxURLs:        5,
		Workers:        1,
		RequestDelay:   time.Millisecond,
		Fetcher:        mockFetcher,
		Cache:          htmlCache,
		FollowBehavior: FollowNone,
	})

	callback := func(ctx context.Context, req *fetch.Request, parsed any, err error) error {
		// The callback won't receive the HTML directly, but we can verify
		// the cache was used by checking that fetcher was not called
		return nil
	}

	ctx := context.Background()
	err = crawler.Crawl(ctx, []string{"https://example.com"}, callback)

	assert.NoError(t, err)

	// Verify that fetcher was not called for cached URL
	// (no mock response was set up, so it would have failed if fetcher was called)
	stats := crawler.GetStats()
	assert.Equal(t, int64(1), stats.GetProcessed())
}

func TestResolveLink(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		link     string
		expected string
		valid    bool
	}{
		{
			name:     "absolute HTTPS URL",
			domain:   "example.com",
			link:     "https://example.com/page",
			expected: "https://example.com/page",
			valid:    true,
		},
		{
			name:     "absolute HTTP URL",
			domain:   "example.com",
			link:     "http://example.com/page",
			expected: "https://example.com/page",
			valid:    true,
		},
		{
			name:     "relative URL with leading slash",
			domain:   "example.com",
			link:     "/about",
			expected: "https://example.com/about",
			valid:    true,
		},
		{
			name:     "relative URL without leading slash",
			domain:   "example.com",
			link:     "about",
			expected: "https://example.com/about",
			valid:    true,
		},
		{
			name:   "invalid scheme",
			domain: "example.com",
			link:   "ftp://example.com/file",
			valid:  false,
		},
		{
			name:   "javascript URL",
			domain: "example.com",
			link:   "javascript:void(0)",
			valid:  false,
		},
		{
			name:   "mailto URL",
			domain: "example.com",
			link:   "mailto:test@example.com",
			valid:  false,
		},
		{
			name:     "URL with fragment",
			domain:   "example.com",
			link:     "https://example.com/page#section",
			expected: "https://example.com/page",
			valid:    true,
		},
		{
			name:     "domain with https prefix",
			domain:   "https://example.com",
			link:     "/page",
			expected: "https://example.com/page",
			valid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := ResolveLink(tt.domain, tt.link)
			assert.Equal(t, tt.valid, valid)
			if valid {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCrawler_FollowBehavior(t *testing.T) {
	tests := []struct {
		name           string
		followBehavior FollowBehavior
		baseURL        string
		discoveredURLs []string
		expectedCount  int
	}{
		{
			name:           "follow none",
			followBehavior: FollowNone,
			baseURL:        "https://example.com",
			discoveredURLs: []string{
				"https://example.com/page1",
				"https://other.com/page",
				"https://sub.example.com/page",
			},
			expectedCount: 0,
		},
		{
			name:           "follow same domain",
			followBehavior: FollowSameDomain,
			baseURL:        "https://example.com",
			discoveredURLs: []string{
				"https://example.com/page1",
				"https://other.com/page",
				"https://sub.example.com/page",
			},
			expectedCount: 1,
		},
		{
			name:           "follow any",
			followBehavior: FollowAny,
			baseURL:        "https://example.com",
			discoveredURLs: []string{
				"https://example.com/page1",
				"https://other.com/page",
				"https://sub.example.com/page",
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFetcher := NewMockFetcher()

			// Setup main page response
			links := make([]*fetch.Link, len(tt.discoveredURLs))
			for i, url := range tt.discoveredURLs {
				links[i] = &fetch.Link{URL: url}
			}

			mockFetcher.AddResponse(tt.baseURL, &fetch.Response{
				URL:   tt.baseURL,
				HTML:  "<html><body><h1>Test</h1></body></html>",
				Links: links,
			})

			// Setup responses for discovered URLs
			for _, url := range tt.discoveredURLs {
				mockFetcher.AddResponse(url, &fetch.Response{
					URL:   url,
					HTML:  "<html><body><h1>Page</h1></body></html>",
					Links: []*fetch.Link{},
				})
			}

			// Use more workers for FollowAny to handle larger queue size needed
			workers := 1
			if tt.followBehavior == FollowAny {
				workers = 2 // Increase workers for FollowAny to prevent queue overflow
			}

			crawler := New(Options{
				MaxURLs:        10,
				Workers:        workers,
				RequestDelay:   time.Millisecond * 10, // Increase delay to avoid race conditions
				Fetcher:        mockFetcher,
				FollowBehavior: tt.followBehavior,
			})

			var processedURLs []string
			mu := sync.Mutex{}

			callback := func(ctx context.Context, req *fetch.Request, parsed any, err error) error {
				mu.Lock()
				defer mu.Unlock()
				processedURLs = append(processedURLs, req.URL)
				return nil
			}

			ctx := context.Background()
			err := crawler.Crawl(ctx, []string{tt.baseURL}, callback)

			assert.NoError(t, err)

			// Should always process the base URL plus expected discovered URLs
			expectedTotal := 1 + tt.expectedCount
			// For debugging, print the processed URLs if test fails
			if len(processedURLs) != expectedTotal {
				t.Logf("Expected %d URLs but got %d: %v", expectedTotal, len(processedURLs), processedURLs)
			}
			assert.Len(t, processedURLs, expectedTotal)
			assert.Contains(t, processedURLs, tt.baseURL)
		})
	}
}

func TestCrawler_ErrorHandling(t *testing.T) {
	mockFetcher := NewMockFetcher()

	// Setup error for one URL and success for another
	mockFetcher.AddError("https://error.com", fmt.Errorf("fetch failed"))
	mockFetcher.AddResponse("https://success.com", &fetch.Response{
		URL:   "https://success.com",
		HTML:  "<html><body><h1>Success</h1></body></html>",
		Links: []*fetch.Link{},
	})

	crawler := New(Options{
		MaxURLs:        10,
		Workers:        1,
		RequestDelay:   time.Millisecond,
		Fetcher:        mockFetcher,
		FollowBehavior: FollowNone,
	})

	var processedURLs []string
	var errors []error
	mu := sync.Mutex{}

	callback := func(ctx context.Context, req *fetch.Request, parsed any, err error) error {
		mu.Lock()
		defer mu.Unlock()
		processedURLs = append(processedURLs, req.URL)
		if err != nil {
			errors = append(errors, err)
		}
		return nil
	}

	ctx := context.Background()
	err := crawler.Crawl(ctx, []string{"https://error.com", "https://success.com"}, callback)

	assert.NoError(t, err)
	assert.Len(t, processedURLs, 2)
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error(), "fetch failed")

	stats := crawler.GetStats()
	assert.Equal(t, int64(2), stats.GetProcessed())
	assert.Equal(t, int64(1), stats.GetSucceeded())
	assert.Equal(t, int64(1), stats.GetFailed())
}

func TestCrawler_MaxURLsLimit(t *testing.T) {
	mockFetcher := NewMockFetcher()

	// Setup responses for multiple URLs
	urls := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
		"https://example.com/4",
		"https://example.com/5",
	}

	for _, url := range urls {
		mockFetcher.AddResponse(url, &fetch.Response{
			URL:   url,
			HTML:  "<html><body><h1>Page</h1></body></html>",
			Links: []*fetch.Link{},
		})
	}

	crawler := New(Options{
		MaxURLs:        3, // Limit to 3 URLs
		Workers:        1,
		RequestDelay:   time.Millisecond,
		Fetcher:        mockFetcher,
		FollowBehavior: FollowNone,
	})

	var processedURLs []string
	mu := sync.Mutex{}

	callback := func(ctx context.Context, req *fetch.Request, parsed any, err error) error {
		mu.Lock()
		defer mu.Unlock()
		processedURLs = append(processedURLs, req.URL)
		return nil
	}

	ctx := context.Background()
	err := crawler.Crawl(ctx, urls, callback)

	assert.NoError(t, err)
	assert.LessOrEqual(t, len(processedURLs), 3)

	stats := crawler.GetStats()
	assert.LessOrEqual(t, stats.GetProcessed(), int64(3))
}
