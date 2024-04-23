package crawler

import (
	"bytes"
	"container/list"
	"context"
	"io"
	"log/slog"
	"net/url"
	"sync"

	"github.com/adrianos93/crawler/internal/parser"
)

// Crawler is used to crawl webpages and populate a linked list with the extracted information
type Crawler struct {
	logger *slog.Logger
	reader Reader
	mw     int
}

// Reader is an interface so that Crawler can read the data provided from the given source
type Reader interface {
	ReadPage(ctx context.Context, location string, data io.Writer) error
}

// New returns a Crawler
func New(logger *slog.Logger, reader Reader, maxWorkers int) *Crawler {
	return &Crawler{
		reader: reader,
		mw:     maxWorkers,
		logger: logger,
	}
}

// Page is a type used to hold the information extracted from a given page
type Page struct {
	URL   string
	Links []string
}

// Start is used to initiate the crawling process.
// It takes a root URL and will trigger a couple of goroutines to crawl the extracted data.
// It will create a buffered channel for the links, with a max buffer provided to the crawler to ensure concurrency is not unbounded.
// Once all the data is extracted it will loop over the linked list on the Crawler and populate and return slice of Page
func (c *Crawler) Start(ctx context.Context, root *url.URL) []Page {
	pages := []Page{}
	pageChan := make(chan Page, c.mw)

	data := list.New()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go c.process(ctx, pageChan, wg, data)
	go c.crawl(ctx, root, pageChan)
	wg.Wait()

	for e := data.Front(); e != nil; e = e.Next() {
		pages = append(pages, e.Value.(Page))
	}
	return pages
}

// crawl will read in the given page and extract the links from the returned io.Writer
// It will append the data to the linked list of the Crawler and send all the links to the links channel
func (c *Crawler) crawl(ctx context.Context, page *url.URL, linksChan chan Page) {
	crawledPage := Page{}
	defer func() {
		linksChan <- crawledPage
	}()
	var b bytes.Buffer
	if err := c.reader.ReadPage(ctx, page.String(), &b); err != nil {
		c.logger.Error("reading page", "error", err, "page", page.String())
		return
	}
	links, err := parser.ParseLinks(page, &b)
	if err != nil {
		c.logger.Error("parsing links", "error", err, "page", page.String())
		return
	}

	crawledPage.URL = page.String()
	if len(links) > 0 {
		for _, link := range links {
			crawledPage.Links = append(crawledPage.Links, link.String())
		}
	}
}

// process is another routine that reads from the links channel and will check if the url has been visited before.
// If the URL has not been seen before, it will trigger another crawl goroutine and increment the wait group.
// A receive from the channel counts as a crawl routine ending, so the wg counter is decremented.
// It also honours context.Context, so before it attempts to receive from the channel, it will check for a cancellation and return early if it has been cancelled.
// Due to having a local hashmap, only one instance of process can run at any given time to avoid locking.
// It will also add the Page to the linked list for further processing.
func (c *Crawler) process(ctx context.Context, nodeChan chan Page, wg *sync.WaitGroup, data *list.List) {
	seen := make(map[string]struct{})
	for {
		select {
		case <-ctx.Done():
			return
		// channel is never closed, so no point in checking for it.
		case node := <-nodeChan:
			data.PushBack(node)
			for _, link := range node.Links {
				if _, ok := seen[link]; !ok {
					seen[link] = struct{}{}

					parsed, err := url.Parse(link)
					if err != nil {
						c.logger.Error("invalid link", "error", err)
						continue
					}

					wg.Add(1)
					go c.crawl(ctx, parsed, nodeChan)
				}
			}
			wg.Done()
		}
	}
}
