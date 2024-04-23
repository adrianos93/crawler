package reader

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

type fakeLocation struct {
	T *testing.T
}

func (f *fakeLocation) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(w, strings.NewReader(testHTML))
}

func Test_ReadPage(t *testing.T) {
	testServer := &fakeLocation{T: t}
	ts := httptest.NewServer(http.HandlerFunc(testServer.ServeHTTP))
	defer ts.Close()
	for name, test := range map[string]struct {
		location string

		want      string
		assertErr require.ErrorAssertionFunc
	}{
		"successfully reads page": {
			location:  ts.URL,
			want:      testHTML,
			assertErr: require.NoError,
		},
		"request fails": {
			location:  "",
			assertErr: require.Error,
		},
	} {
		t.Run(name, func(t *testing.T) {
			r := New(1 * time.Second)

			var b bytes.Buffer
			err := r.ReadPage(context.Background(), test.location, &b)
			test.assertErr(t, err)
			require.Equal(t, test.want, b.String())
		})
	}
}
