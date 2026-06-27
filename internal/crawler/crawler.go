package crawler

import (
	"context"
	"io"
	"log/slog"
	"net/url"
	"sort"

	"github.com/adrianos93/crawler/internal/parser"
)

// Crawler is used to crawl webpages and collect the links extracted from each one.
type Crawler struct {
	logger  *slog.Logger
	reader  Reader
	workers int
}

// Reader fetches the content of a page as a stream. The caller closes the returned reader.
type Reader interface {
	Fetch(ctx context.Context, location string) (io.ReadCloser, error)
}

// New returns a Crawler. workers bounds the number of pages fetched concurrently
// and is clamped to at least 1.
func New(logger *slog.Logger, reader Reader, workers int) *Crawler {
	if workers < 1 {
		workers = 1
	}
	return &Crawler{
		reader:  reader,
		workers: workers,
		logger:  logger,
	}
}

// Page holds a crawled page and the in-domain links discovered on it.
type Page struct {
	URL   string
	Links []string
}

// Start crawls every reachable in-domain page beginning at root and returns the
// collected pages sorted by URL.
//
// A single coordinator goroutine owns the visited set and result slice, so no
// locking is required. Each discovered URL is crawled in its own goroutine, but a
// semaphore bounds how many fetches run at once to c.workers. The coordinator
// tracks outstanding crawls explicitly and returns as soon as the queue drains or
// ctx is cancelled, leaving no goroutines behind.
func (c *Crawler) Start(ctx context.Context, root *url.URL) []Page {
	// Buffer results so finished crawls don't block waiting for the coordinator
	// to catch up between dispatching new work.
	results := make(chan Page, c.workers)
	sem := make(chan struct{}, c.workers)

	seen := map[string]struct{}{root.String(): {}}
	pages := []Page{}
	outstanding := 0

	enqueue := func(u *url.URL) {
		outstanding++
		go c.crawl(ctx, u, sem, results)
	}

	enqueue(root)

	for outstanding > 0 {
		var page Page
		select {
		case <-ctx.Done():
			// Abandon in-flight crawls; they observe ctx.Done() on their send and exit.
			return sortPages(pages)
		case page = <-results:
		}
		outstanding--

		if page.URL != "" {
			pages = append(pages, page)
		}

		for _, link := range page.Links {
			if _, ok := seen[link]; ok {
				continue
			}
			seen[link] = struct{}{}

			parsed, err := url.Parse(link)
			if err != nil {
				c.logger.Error("invalid link", "error", err, "link", link)
				continue
			}
			enqueue(parsed)
		}
	}

	return sortPages(pages)
}

// crawl fetches and parses a single page, then sends exactly one Page on results so
// the coordinator can account for it. On failure it sends a zero-value Page (empty
// URL), which the coordinator drops. The send honours ctx so a cancelled crawl never
// blocks forever.
func (c *Crawler) crawl(ctx context.Context, u *url.URL, sem chan struct{}, results chan<- Page) {
	page := Page{}
	defer func() {
		select {
		case results <- page:
		case <-ctx.Done():
		}
	}()

	// Bound the number of concurrent fetches.
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-ctx.Done():
		return
	}

	body, err := c.reader.Fetch(ctx, u.String())
	if err != nil {
		c.logger.Warn("fetching page", "error", err, "page", u.String())
		return
	}
	defer body.Close()

	links, err := parser.ParseLinks(u, body)
	if err != nil {
		c.logger.Error("parsing links", "error", err, "page", u.String())
		return
	}

	page.URL = u.String()
	for _, link := range links {
		page.Links = append(page.Links, link.String())
	}
}

func sortPages(pages []Page) []Page {
	sort.Slice(pages, func(i, j int) bool { return pages[i].URL < pages[j].URL })
	return pages
}
