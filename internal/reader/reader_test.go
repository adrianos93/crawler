package reader

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testHTML = `<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
</head>
<body>

<h1>Just a text header</h1>
<p><a href=/next>Next page</a></p>

<p><a href=localhost/about>More page</a></p>
</body>
</html>`

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, testHTML)
	})
	mux.HandleFunc("/notfound", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/pdf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = io.WriteString(w, "%PDF-1.4 not html")
	})
	return httptest.NewServer(mux)
}

func Test_Fetch(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()

	for name, test := range map[string]struct {
		location  string
		want      string
		assertErr require.ErrorAssertionFunc
	}{
		"successfully reads HTML page": {
			location:  ts.URL + "/ok",
			want:      testHTML,
			assertErr: require.NoError,
		},
		"rejects non-2xx status": {
			location:  ts.URL + "/notfound",
			assertErr: require.Error,
		},
		"rejects non-HTML content type": {
			location:  ts.URL + "/pdf",
			assertErr: require.Error,
		},
		"request fails on empty location": {
			location:  "",
			assertErr: require.Error,
		},
	} {
		t.Run(name, func(t *testing.T) {
			r := New(1 * time.Second)

			body, err := r.Fetch(context.Background(), test.location)
			test.assertErr(t, err)
			if err != nil {
				require.Nil(t, body)
				return
			}
			defer body.Close()

			got, err := io.ReadAll(body)
			require.NoError(t, err)
			require.Equal(t, test.want, string(got))
		})
	}
}
