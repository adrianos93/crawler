package crawler

import (
	"context"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testData = `<!DOCTYPE html>
	<html>
	<head>
	<title>Page Title</title>
	</head>
	<body>
	
	<h1>Just a text header</h1>
	
	<p><a href=/next>Next page</a></p>
	</body>
	</html>
`

type mockReader struct {
	mock.Mock
}

func (mr *mockReader) ReadPage(ctx context.Context, location string, data io.Writer) error {
	if strings.Contains(location, "/next") {
		_, _ = io.Copy(data, strings.NewReader(""))
		return mr.Called(ctx, location, data).Error(0)
	}
	_, _ = io.Copy(data, strings.NewReader(testData))
	return mr.Called(ctx, location, data).Error(0)
}

func Test_Start(t *testing.T) {
	for name, test := range map[string]struct {
		root *url.URL

		readPageFunc func(*mockReader)

		want []Page
	}{
		"returns 2 pages, root with 1 links, and the other link childless": {
			root: urlMaker("http://localhost/"),
			readPageFunc: func(mr *mockReader) {
				mr.On("ReadPage", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			want: []Page{
				{URL: "http://localhost/", Links: []string{"http://localhost/next"}},
				{URL: "http://localhost/next"},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			readerMock := &mockReader{}
			readerMock.Test(t)
			defer readerMock.AssertExpectations(t)

			if test.readPageFunc != nil {
				test.readPageFunc(readerMock)
			}

			crawler := New(slog.Default(), readerMock, 5)
			got := crawler.Start(context.Background(), test.root)
			require.Equal(t, test.want, got)
		})
	}
}

func urlMaker(in string) *url.URL {
	parsed, _ := url.Parse(in)
	return parsed
}
