package crawler

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/myzie/web"
	"github.com/myzie/web/cache"
	"github.com/myzie/web/fetch"
)

// FollowBehavior is used to determine how to follow links.
type FollowBehavior string

const (
	FollowAny               FollowBehavior = "any"
	FollowSameDomain        FollowBehavior = "same_domain"
	FollowRelatedSubdomains FollowBehavior = "related_subdomains"
	FollowNone              FollowBehavior = "none"
)

// Parser is an interface describing a webpage parser. It accepts the fetched
// page and returns a parsed object.
type Parser interface {
	Parse(ctx context.Context, page *fetch.Response) (any, error)
}

// ProcessCallback is called with the fetch request and parsed result (if any)
type Callback func(ctx context.Context, req *fetch.Request, parsed any, err error)

// Options used to configure a crawler.
type Options struct {
	MaxURLs              int
	Workers              int
	Cache                cache.Cache
	Fetcher              fetch.Fetcher
	FetcherName          string
	RequestDelay         time.Duration
	KnownURLs            []string
	Parsers              map[string]Parser
	DefaultParser        Parser
	FollowBehavior       FollowBehavior
	Logger               *slog.Logger
	ShowProgress         bool
	ShowProgressInterval time.Duration
	QueueSize            int
}

// Crawler is used to crawl the web.
type Crawler struct {
	processedURLs        sync.Map
	queue                chan string
	maxURLs              int
	workers              int
	requestDelay         time.Duration
	cache                cache.Cache
	fetcher              fetch.Fetcher
	fetcherName          string
	knownURLs            []string
	parsers              map[string]Parser
	defaultParser        Parser
	followBehavior       FollowBehavior
	activeWorkers        int64
	stats                *CrawlerStats
	logger               *slog.Logger
	running              bool
	showProgress         bool
	showProgressInterval time.Duration
}

// New creates a new crawler.
func New(opts Options) *Crawler {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	if opts.ShowProgress && opts.ShowProgressInterval == 0 {
		opts.ShowProgressInterval = 30 * time.Second
	}
	if opts.QueueSize <= 0 {
		opts.QueueSize = 10000
	}
	return &Crawler{
		cache:                opts.Cache,
		maxURLs:              opts.MaxURLs,
		workers:              opts.Workers,
		requestDelay:         opts.RequestDelay,
		fetcher:              opts.Fetcher,
		fetcherName:          opts.FetcherName,
		knownURLs:            opts.KnownURLs,
		parsers:              opts.Parsers,
		followBehavior:       opts.FollowBehavior,
		defaultParser:        opts.DefaultParser,
		stats:                &CrawlerStats{},
		logger:               logger,
		showProgress:         opts.ShowProgress,
		showProgressInterval: opts.ShowProgressInterval,
		queue:                make(chan string, opts.QueueSize),
	}
}

// incrementActiveWorkers atomically increments the active workers counter
func (c *Crawler) incrementActiveWorkers() {
	atomic.AddInt64(&c.activeWorkers, 1)
}

// decrementActiveWorkers atomically decrements the active workers counter
func (c *Crawler) decrementActiveWorkers() {
	atomic.AddInt64(&c.activeWorkers, -1)
}

// getActiveWorkers atomically gets the current active workers count
func (c *Crawler) getActiveWorkers() int64 {
	return atomic.LoadInt64(&c.activeWorkers)
}

func (c *Crawler) getFetcherName() string {
	if c.fetcherName != "" {
		return c.fetcherName
	}
	return "http"
}

// Crawl the provided URLs and call the callback for each processed page.
// Links may be followed depending on the configured follow behavior.
func (c *Crawler) Crawl(ctx context.Context, urls []string, callback Callback) error {
	if c.running {
		return errors.New("crawler is already running")
	}
	c.running = true

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		c.running = false
		cancel()
	}()

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < c.workers; i++ {
		wg.Add(1)
		go c.worker(ctx, i, &wg, callback)
	}
	defer close(c.queue)

	// Start progress reporter
	if c.showProgress {
		go c.progressReporter(ctx)
	}

	// Start idle monitor to detect when no more work is available
	go c.idleMonitor(ctx, cancel)

	// Queue initial URLs
	count, err := c.enqueue(ctx, urls)
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	// Wait for workers to complete
	wg.Wait()
	return nil
}

func (c *Crawler) enqueue(ctx context.Context, urls []string) (int, error) {
	queued := 0
	for _, rawURL := range urls {
		url, err := web.NormalizeURL(rawURL)
		if err != nil {
			return queued, err
		}
		value := url.String()
		if _, exists := c.processedURLs.LoadOrStore(value, true); !exists {
			select {
			case c.queue <- value:
				queued++
			case <-ctx.Done():
				return queued, ctx.Err()
			}
		}
	}
	return queued, nil
}

func (c *Crawler) worker(ctx context.Context, workerID int, wg *sync.WaitGroup, callback Callback) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case rawURL, ok := <-c.queue:
			if !ok {
				return
			}
			if c.stats.GetProcessed() >= int64(c.maxURLs) {
				return
			}
			c.incrementActiveWorkers()
			c.processURL(ctx, rawURL, callback)
			c.decrementActiveWorkers()
			if c.requestDelay > 0 {
				time.Sleep(c.requestDelay)
			}
		}
	}
}

func (c *Crawler) processURL(ctx context.Context, rawURL string, callback Callback) {
	c.stats.IncrementProcessed()

	// Parse the url to get its domain
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		c.logger.Warn("invalid url",
			slog.String("url", rawURL),
			slog.String("error", err.Error()))
		return
	}
	domain := parsedURL.Hostname()

	// Check cache first if one is enabled
	var response *fetch.Response
	if c.cache != nil {
		if cachedHTML, err := c.cache.Get(ctx, rawURL); err == nil {
			c.logger.Debug("cache hit", slog.String("url", rawURL))
			response = &fetch.Response{
				URL:  rawURL,
				HTML: string(cachedHTML),
			}
		}
	}

	// Create fetch request
	req := &fetch.Request{
		URL:             rawURL,
		Prettify:        false,
		OnlyMainContent: false,
		Fetcher:         c.getFetcherName(),
	}

	// Fetch if there was not a cache hit
	if response == nil {
		c.logger.Debug("fetching", slog.String("url", rawURL))
		response, err = c.fetcher.Fetch(ctx, req)
		if err != nil {
			callback(ctx, req, nil, err)
			c.stats.IncrementFailed()
			return
		}
		if c.cache != nil && response.HTML != "" {
			if err := c.cache.Set(ctx, rawURL, []byte(response.HTML)); err != nil {
				c.logger.Warn("failed to cache html",
					slog.String("url", rawURL),
					slog.String("error", err.Error()))
			}
		}
	}

	// Parse if a parser exists for the domain
	var parsed any
	var discoveredURLs []string
	var parseErr error

	parser, exists := c.getParser(domain)
	if exists {
		c.logger.Info("parsing with domain parser",
			slog.String("url", rawURL),
			slog.String("domain", domain))
		parsed, parseErr = parser.Parse(ctx, response)
		if parseErr != nil {
			c.logger.Error("failed to parse",
				slog.String("url", rawURL),
				slog.String("error", parseErr.Error()))
		}
	}

	// Extract URLs from the page
	if response.Links != nil {
		discoveredURLs = c.extractURLs(response.Links, domain)
	}

	callback(ctx, req, parsed, parseErr)
	filteredURLs := c.filterURLs(parsedURL, discoveredURLs)
	c.queueDiscoveredURLs(ctx, filteredURLs)
	c.stats.IncrementSucceeded()
}

func (c *Crawler) getParser(domain string) (Parser, bool) {
	if parser, exists := c.parsers[domain]; exists {
		return parser, true
	}
	if c.defaultParser != nil {
		return c.defaultParser, true
	}
	return nil, false
}

func (c *Crawler) filterURLs(pageURL *url.URL, links []string) []string {
	if c.followBehavior == FollowNone {
		return nil
	}
	var filtered []string
	for _, rawURL := range links {
		u, err := web.NormalizeURL(rawURL)
		if err != nil {
			continue
		}
		switch c.followBehavior {
		case FollowAny:
			filtered = append(filtered, rawURL)
		case FollowSameDomain:
			if web.AreSameHost(u, pageURL) {
				filtered = append(filtered, rawURL)
			}
		case FollowRelatedSubdomains:
			if web.AreRelatedHosts(u, pageURL) {
				filtered = append(filtered, rawURL)
			}
		}
	}
	return filtered
}

func (c *Crawler) extractURLs(links []*fetch.Link, domain string) []string {
	urlMap := make(map[string]bool)
	for _, link := range links {
		if url, ok := ResolveLink(domain, link.URL); ok {
			urlMap[url] = true
		}
	}
	var results []string
	for url := range urlMap {
		results = append(results, url)
	}
	sort.Strings(results)
	return results
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
		normalizedURL, err := web.NormalizeURL(parsedURL.String())
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
	normalizedURL, err := web.NormalizeURL(resolvedURL.String())
	if err != nil {
		return "", false
	}
	return normalizedURL.String(), true
}

func (c *Crawler) queueDiscoveredURLs(ctx context.Context, urls []string) {
	var next []string
	for _, rawURL := range urls {
		if c.stats.GetProcessed() >= int64(c.maxURLs) {
			return
		}
		u, err := web.NormalizeURL(rawURL)
		if err != nil {
			c.logger.Warn("invalid url",
				slog.String("url", rawURL),
				slog.String("error", err.Error()))
			continue
		}
		rawURL = u.String()
		if _, exists := c.processedURLs.LoadOrStore(rawURL, true); !exists {
			next = append(next, rawURL)
		}
	}

	select {
	case c.queue <- urlStr:
	case <-ctx.Done():
		return
	default:
		// Queue is full, skip this URL
	}
}

func (c *Crawler) progressReporter(ctx context.Context) {
	ticker := time.NewTicker(c.showProgressInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.logger.Info("crawl progress",
				slog.Int64("processed", c.stats.GetProcessed()),
				slog.Int64("succeeded", c.stats.GetSucceeded()),
				slog.Int64("failed", c.stats.GetFailed()))
		}
	}
}

// GetStats returns the current crawling statistics
func (c *Crawler) GetStats() *CrawlerStats {
	return c.stats
}

func (c *Crawler) idleMonitor(ctx context.Context, cancel context.CancelFunc) {
	// Check every second for idle state
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check if we're idle: no active workers and queue is empty
			if c.getActiveWorkers() == 0 && len(c.queue) == 0 {
				c.logger.Info("no more work available, stopping crawler")
				cancel() // Cancel context to stop all workers
				return
			}
		}
	}
}
