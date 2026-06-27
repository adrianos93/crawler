package reader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// maxBodySize caps how much of a response body we will read into the parser,
	// protecting against accidentally pulling huge or malicious payloads into memory.
	maxBodySize = 10 << 20 // 10 MiB

	// userAgent identifies the crawler. Many sites reject the default Go user agent.
	userAgent = "adrianos93-crawler/1.0 (+https://github.com/adrianos93/crawler)"
)

// Reader is used to connect to webpages and read their content
type Reader struct {
	client http.Client
}

// New returns a new Reader
func New(timeout time.Duration) Reader {
	return Reader{
		client: http.Client{
			Timeout: timeout,
		},
	}
}

// limitedBody pairs a size-limited reader with the underlying body's Closer so the
// caller can stream the response and still release the connection.
type limitedBody struct {
	io.Reader
	io.Closer
}

// Fetch performs a GET against location and returns the response body as a stream.
// The caller is responsible for closing the returned ReadCloser.
//
// Non-2xx responses and non-HTML content types are rejected so the parser only ever
// sees HTML it can meaningfully process. The body is capped at maxBodySize.
func (r Reader) Fetch(ctx context.Context, location string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return nil, fmt.Errorf("generating request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting data: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, location)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "text/html") {
		resp.Body.Close()
		return nil, fmt.Errorf("skipping non-HTML content type %q for %s", ct, location)
	}

	return limitedBody{
		Reader: io.LimitReader(resp.Body, maxBodySize),
		Closer: resp.Body,
	}, nil
}
