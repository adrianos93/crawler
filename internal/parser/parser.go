package parser

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// regex to match file extensions
var extensionRegEx = regexp.MustCompile(`\.[a-zA-Z0-9]+$`)

// ParseLinks takes a given URL and will attempt to extract the links found in the data provide via the io.Reader.
// The domain URL is used to filter out links not associated with it
// Valid links are converted to a *url.URL and returned as part of a slice
// Invalid links are discarded
// A hashmap is used to avoid appending duplicate links to the returned slice.
func ParseLinks(domain *url.URL, r io.Reader) ([]*url.URL, error) {
	tokenizer := html.NewTokenizer(r)
	links := []*url.URL{}
	seen := map[string]struct{}{}
	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				break
			}
			return nil, fmt.Errorf("extracting links: %w", tokenizer.Err())
		}

		token := tokenizer.Token()
		if isAnchorTag(tokenType, token) {
			for _, attr := range token.Attr {
				if attr.Key == atom.Href.String() {
					link := verifyAndFormatLink(domain, attr.Val)
					if link != nil {
						_, ok := seen[link.String()]
						if !ok {
							seen[link.String()] = struct{}{}
							links = append(links, link)
						}
						continue
					}

				}
			}
		}
	}
	return links, nil
}

func verifyAndFormatLink(page *url.URL, link string) *url.URL {
	parsedLink, err := url.Parse(link)
	if err != nil {
		return nil
	}

	// do not include file paths
	if strings.Contains(parsedLink.Path, "..") {
		return nil
	}

	// avoid files (.pdf, .jpg)
	if extensionRegEx.MatchString(parsedLink.Path) {
		return nil
	}

	// fragments refer to content on the same page so we want to remove them
	if parsedLink.Fragment != "" {
		parsedLink.Fragment = ""
	}

	// empty host but populated path means it's a path on the same domain
	if parsedLink.Host == "" && (parsedLink.Path != "" && parsedLink.Path != "/") {
		parsedLink.Scheme = page.Scheme
		parsedLink.Host = page.Host

		return parsedLink
	}

	// same host and the path is not empty or has a trailing slash
	if parsedLink.Host == page.Host && (parsedLink.Path != "" && parsedLink.Path != "/") {
		return parsedLink
	}

	return nil
}

func isAnchorTag(tokenType html.TokenType, token html.Token) bool {
	return tokenType == html.StartTagToken && token.DataAtom == atom.A
}
