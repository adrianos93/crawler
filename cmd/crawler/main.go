package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adrianos93/crawler/internal/crawler"
	"github.com/adrianos93/crawler/internal/reader"
)

var (
	rootURL     string
	concurrency int
	timeout     time.Duration
)

func init() {
	flag.StringVar(&rootURL, "domain", "", "domain to crawl (e.g. https://example.com)")
	flag.IntVar(&concurrency, "concurrency", 5, "number of pages to crawl concurrently")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "per-request HTTP timeout")
}

func main() {
	flag.Parse()

	// Logs go to stderr so stdout carries only crawl results and stays pipeable.
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	if concurrency < 1 {
		logger.Error("concurrency must be at least 1")
		os.Exit(1)
	}

	domain, err := url.Parse(rootURL)
	if err != nil || domain.Scheme == "" || domain.Host == "" {
		logger.Error("domain must be an absolute URL with a scheme and host, e.g. https://example.com", "domain", rootURL)
		os.Exit(1)
	}

	// Cancel the crawl on the first interrupt so in-flight work terminates gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	crawl := crawler.New(logger, reader.New(timeout), concurrency)

	pages := crawl.Start(ctx, domain)

	for _, page := range pages {
		fmt.Printf("Page %s has links: %v\n", page.URL, page.Links)
	}
}
