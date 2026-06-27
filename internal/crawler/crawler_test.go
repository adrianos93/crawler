package crawler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// stubReader serves canned HTML per location and optional per-location errors.
type stubReader struct {
	pages map[string]string
	errs  map[string]error
}

func (s *stubReader) Fetch(ctx context.Context, location string) (io.ReadCloser, error) {
	if err := s.errs[location]; err != nil {
		return nil, err
	}
	body, ok := s.pages[location]
	if !ok {
		return nil, fmt.Errorf("not found: %s", location)
	}
	return io.NopCloser(strings.NewReader(body)), nil
}

func link(path string) string { return fmt.Sprintf(`<p><a href=%s>link</a></p>`, path) }

func Test_Start(t *testing.T) {
	for name, test := range map[string]struct {
		reader *stubReader
		want   []Page
	}{
		"crawls reachable pages and de-duplicates links across pages": {
			reader: &stubReader{
				pages: map[string]string{
					"http://localhost/":  link("/a") + link("/b"),
					"http://localhost/a": link("/b"), // /b already discovered from root
					"http://localhost/b": "",
				},
			},
			want: []Page{
				{URL: "http://localhost/", Links: []string{"http://localhost/a", "http://localhost/b"}},
				{URL: "http://localhost/a", Links: []string{"http://localhost/b"}},
				{URL: "http://localhost/b"},
			},
		},
		"drops pages that fail to fetch but keeps crawling the rest": {
			reader: &stubReader{
				pages: map[string]string{
					"http://localhost/":  link("/a") + link("/b"),
					"http://localhost/b": "",
				},
				errs: map[string]error{
					"http://localhost/a": fmt.Errorf("boom"),
				},
			},
			want: []Page{
				{URL: "http://localhost/", Links: []string{"http://localhost/a", "http://localhost/b"}},
				{URL: "http://localhost/b"},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := New(slog.Default(), test.reader, 5)
			got := c.Start(context.Background(), urlMaker("http://localhost/"))
			require.Equal(t, test.want, got)
		})
	}
}

// Test_Start_Cancelled ensures a cancelled context makes Start return promptly
// without deadlocking or leaking goroutines.
func Test_Start_Cancelled(t *testing.T) {
	reader := &stubReader{pages: map[string]string{"http://localhost/": link("/a")}}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := New(slog.Default(), reader, 5)
	got := c.Start(ctx, urlMaker("http://localhost/"))

	// Cancellation may race the root crawl, so we only require that it returns.
	require.LessOrEqual(t, len(got), 1)
}

func urlMaker(in string) *url.URL {
	parsed, _ := url.Parse(in)
	return parsed
}
