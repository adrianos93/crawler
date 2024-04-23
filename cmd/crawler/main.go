package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/adrianos93/crawler/internal/crawler"
	"github.com/adrianos93/crawler/internal/reader"
)

var (
	rootURL     string
	concurrency int
)

func init() {
	flag.StringVar(&rootURL, "domain", "", "domain to crawl")
	flag.IntVar(&concurrency, "concurrency", 5, "the amount of concurrent workers to spawn")
}

func main() {
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if rootURL == "" {
		logger.Error("a valid URL for crawling must be provided")
		os.Exit(1)

		if !strings.Contains(rootURL, "http://") || !strings.Contains(rootURL, "https://") {
			logger.Error("url must contain a valid protocol. Either http or https")
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	newReader := reader.New(5 * time.Second)

	// this method will notify the downstream processes of a cancellation so that the work can be gracefully terminated
	go func() {
		c := make(chan os.Signal, 1)

		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()

	domain, err := url.Parse(rootURL)
	if err != nil {
		logger.Error("parsing domain", "error", err)
		os.Exit(1)
	}

	crawl := crawler.New(logger, newReader, concurrency)

	pages := crawl.Start(ctx, domain)

	for i, page := range pages {
		fmt.Printf("Page %s has links: %v\n", page.URL, page.Links)
		if i == 5 {
			break
		}
	}
}
