package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/deepnoodle-ai/web"
	"github.com/deepnoodle-ai/web/crawler"
	"github.com/deepnoodle-ai/web/fetch"
)

func normalize(url string) string {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}
	return url
}

func main() {
	// Parse command line flags
	var (
		urls         = flag.String("urls", "", "Comma-separated list of URLs to crawl")
		inputFile    = flag.String("file", "", "File containing URLs to crawl")
		maxURLs      = flag.Int("max-urls", 100, "Maximum number of URLs to crawl")
		workers      = flag.Int("workers", 5, "Number of concurrent workers")
		timeout      = flag.Duration("timeout", 30*time.Second, "Fetch timeout")
		followMode   = flag.String("follow", "same-domain", "Link following behavior: any, same-domain, related-subdomains, none")
		verbose      = flag.Bool("verbose", false, "Enable verbose logging")
		showProgress = flag.Bool("progress", true, "Show progress updates")
		delay        = flag.Duration("delay", 0, "Delay between requests")
	)
	flag.Parse()

	if *urls == "" && *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -urls or -file flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Configure logging
	var logger *slog.Logger
	if *verbose {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	// Parse target URLs
	var startURLs []string

	if *urls != "" {
		startURLs = strings.Split(*urls, ",")
		for _, url := range startURLs {
			startURLs = append(startURLs, normalize(url))
		}
	}

	if *inputFile != "" {
		items, err := web.ReadFileItems(*inputFile)
		if err != nil {
			log.Fatalf("Failed to read input file: %v", err)
		}
		for _, url := range items {
			startURLs = append(startURLs, normalize(url))
		}
	}

	// Parse follow behavior
	var followBehavior crawler.FollowBehavior
	switch *followMode {
	case "any":
		followBehavior = crawler.FollowAny
	case "same-domain":
		followBehavior = crawler.FollowSameDomain
	case "related-subdomains":
		followBehavior = crawler.FollowRelatedSubdomains
	case "none":
		followBehavior = crawler.FollowNone
	default:
		log.Fatalf("Invalid follow mode: %s", *followMode)
	}

	// Create default fetcher with timeout
	defaultFetcher := fetch.NewHTTPFetcher(fetch.HTTPFetcherOptions{
		Timeout: *timeout,
		Headers: fetch.FakeHeaders,
	})

	// Create crawler
	c, err := crawler.New(crawler.Options{
		MaxURLs:        *maxURLs,
		Workers:        *workers,
		RequestDelay:   *delay,
		DefaultFetcher: defaultFetcher,
		FollowBehavior: followBehavior,
		Logger:         logger,
		ShowProgress:   *showProgress,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}

	// Start crawling
	ctx := context.Background()
	startTime := time.Now()

	err = c.Crawl(ctx, startURLs, func(ctx context.Context, result *crawler.Result) {
		if result.Error != nil {
			logger.Error("Failed to crawl",
				slog.String("url", result.URL.String()),
				slog.String("error", result.Error.Error()))
			return
		}
		logger.Info("Crawled",
			slog.String("url", result.URL.String()),
			slog.Int("links", len(result.Links)),
			slog.Int("status", result.Response.StatusCode))
	})
	if err != nil {
		log.Fatalf("Crawling failed: %v", err)
	}

	// Print final statistics
	stats := c.GetStats()
	duration := time.Since(startTime)
	crawledCount := stats.GetProcessed()
	fmt.Printf("\nCrawling completed in %v\n", duration)
	fmt.Printf("Total URLs processed: %d\n", crawledCount)
	fmt.Printf("Successful: %d\n", stats.GetSucceeded())
	fmt.Printf("Failed: %d\n", stats.GetFailed())
	fmt.Printf("Average rate: %.2f pages/second\n", float64(crawledCount)/duration.Seconds())
}
