# Crawler

Crawler is a simple web crawler expressed as a CLI tool written in Go.

It will attempt to crawl a given domain and extract all the links found on that domain. It will then follow those links and extract the information from them as well, until it has retrieved them all.

Crawler will only parse and follow links associated with the given domain.

For example, if the domain is `https://example.com`, then it would only parse links that are part of the `example.com` website. It will ignore external links, such as `https://google.com` or even prefixed domains like `https://app.example.com`.

## Requirements

In order to use Crawler, a valid domain must be provided. This can be any valid website, such as `https://google.com` or `https://www.wikipedia.org/`.

Crawler was built using [Go 1.22](https://tip.golang.org/doc/go1.22), so in order to run or build the application, users must have the Go programming language installed.

Instructions on how to install Go can be found [here](https://go.dev/doc/install).

## Using the CLI

Once you've selected a domain to crawl and have installed Go, Crawler can be used by just running the program from the root of the repo.

Crawler takes two arguments:

* --domain - this is mandatory and represents the domain you wish Crawler to crawl. Format: `https://example.com`
* --concurrency - this represents the amount of concurrent workers you wish Crawler to spawn. The more workers, the faster it runs, but it will consume more resources. This argument is a simple integer. Default is 5.

```bash
go run cmd/crawler/main.go --domain "<CHOSEN_DOMAIN>"
```

Example with concurrency

```bash
go run cmd/crawler/main.go --domain "<CHOSEN_DOMAIN>" --concurrency 10
```

Output:

```txt
Page https://go.dev/ has links: [https://go.dev/solutions/case-studies https://go.dev/solutions/use-cases https://go.dev/security/ https://go.dev/learn/ https://go.dev/doc/effective_go https://go.dev/doc https://go.dev/doc/devel/release https://go.dev/talks/ https://go.dev/wiki/Conferences https://go.dev/blog https://go.dev/help https://go.dev/dl https://go.dev/dl/ https://go.dev/solutions/ https://go.dev/solutions/google/ https://go.dev/solutions/paypal https://go.dev/solutions/americanexpress https://go.dev/solutions/mercadolibre https://go.dev/tour/ https://go.dev/solutions/cloud/ https://go.dev/solutions/clis/ https://go.dev/pkg/net/http/ https://go.dev/pkg/html/template/ https://go.dev/pkg/database/sql/ https://go.dev/solutions/webdev/ https://go.dev/solutions/devops/ https://go.dev/doc/install/ https://go.dev/learn https://go.dev/play https://go.dev/help/ https://go.dev/pkg/ https://go.dev/project https://go.dev/blog/ https://go.dev/brand https://go.dev/conduct https://go.dev/copyright https://go.dev/tos https://go.dev/s/website-issue]

Page https://go.dev/tos has links: [https://go.dev/solutions/case-studies https://go.dev/solutions/use-cases https://go.dev/security/ https://go.dev/learn/ https://go.dev/doc/effective_go https://go.dev/doc https://go.dev/doc/devel/release https://go.dev/talks/ https://go.dev/wiki/Conferences https://go.dev/blog https://go.dev/help https://go.dev/solutions/ https://go.dev/play https://go.dev/tour/ https://go.dev/help/ https://go.dev/pkg/ https://go.dev/project https://go.dev/dl/ https://go.dev/blog/ https://go.dev/brand https://go.dev/conduct https://go.dev/copyright https://go.dev/tos https://go.dev/s/website-issue]
```
