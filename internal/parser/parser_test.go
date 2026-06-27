package parser

import (
	"net/url"
	"strings"
	"testing"

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
<p><a href=/next>Next page</a></p>
<p><a href=http://localhost/about>More page</a></p>
<p><a href=http://localhost/page.html>HTML page</a></p>
<p><a href=http://localhost/file.pdf>More page</a></p>
<p><a href=http://localhost/terms#go-home>More page</a></p>
<p><a href=http://otherlink.com/what>Nooooo</a></p>
<p><a href=../about>Nooooo</a></p>
</body>
</html>`

func Test_ParseLinks(t *testing.T) {
	for name, test := range map[string]struct {
		domain *url.URL
		data   string

		want      []*url.URL
		assertErr require.ErrorAssertionFunc
	}{
		"keeps in-domain HTML links, drops binaries, fragments, duplicates and external links": {
			domain: urlMaker("http://localhost"),
			data:   testHTML,
			want: func() []*url.URL {
				urls := []*url.URL{}
				// /file.pdf is dropped (binary), /page.html is kept (HTML),
				// /terms#go-home loses its fragment, the rest are external/duplicate/traversal.
				links := []string{
					"http://localhost/next",
					"http://localhost/about",
					"http://localhost/page.html",
					"http://localhost/terms",
				}
				for _, link := range links {
					urls = append(urls, urlMaker(link))
				}
				return urls
			}(),
			assertErr: require.NoError,
		},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := ParseLinks(test.domain, strings.NewReader(test.data))
			test.assertErr(t, err)
			require.Equal(t, test.want, got)
		})

	}
}

func urlMaker(in string) *url.URL {
	parsed, _ := url.Parse(in)
	return parsed
}
