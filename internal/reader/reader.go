package reader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
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

// ReadPage takes a URL and will attempt to connect to it. Resulting data will be written to the provided io.Writer.
func (r Reader) ReadPage(ctx context.Context, location string, data io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return fmt.Errorf("generating request: %w", err)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("requesting data: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(data, resp.Body)
	if err != nil {
		return fmt.Errorf("reading data: %w", err)
	}
	return nil
}
