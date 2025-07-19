package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/myzie/web/crawler"
	"github.com/myzie/web/fetch"
)

type Config struct {
	MaxURLs        int
	Workers        int
	Timeout        time.Duration
	URLs           string
	FollowBehavior string
}

func main() {
	var cfg Config
	flag.IntVar(&cfg.MaxURLs, "max-urls", 100, "maximum number of URLs to crawl")
	flag.IntVar(&cfg.Workers, "workers", 1, "number of workers to use")
	flag.DurationVar(&cfg.Timeout, "timeout", 10*time.Second, "timeout for the fetcher")
	flag.StringVar(&cfg.URLs, "urls", "", "comma separated list of URLs to crawl")
	flag.StringVar(&cfg.FollowBehavior, "follow-behavior", "same-domain", "follow behavior")
	flag.Parse()

	urls := strings.Split(cfg.URLs, ",")
	if len(urls) == 0 {
		log.Fatal("no urls provided")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	fetcher := fetch.NewHTTPFetcher(fetch.HTTPFetcherOptions{
		Timeout:     cfg.Timeout,
		MaxBodySize: 10 * 1024 * 1024,
	})

	c := crawler.New(crawler.Options{
		MaxURLs:        cfg.MaxURLs,
		Workers:        cfg.Workers,
		Fetcher:        fetcher,
		Logger:         logger,
		ShowProgress:   true,
		FollowBehavior: crawler.FollowBehavior(cfg.FollowBehavior),
	})

	callback := func(ctx context.Context, result *crawler.Result) {
		if result.Error != nil {
			logger.Error("error fetching", "url", result.URL, "error", result.Error)
			return
		}
		logger.Info("crawled", "url", result.URL)
	}

	if err := c.Crawl(context.Background(), urls, callback); err != nil {
		log.Fatal(err)
	}
}
