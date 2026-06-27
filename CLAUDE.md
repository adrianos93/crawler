# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the crawler
go run cmd/crawler/main.go --domain "https://example.com"
go run cmd/crawler/main.go --domain "https://example.com" --concurrency 10 --timeout 10s
# --domain is required and must be absolute (scheme+host); concurrency defaults to 5, timeout to 5s.

# Run all tests (use -race; the crawler is concurrent)
go test -race ./...

# Run tests for a specific package
go test -race ./internal/crawler/...

# Build the binary
go build -o crawler cmd/crawler/main.go
```

## Architecture

The crawler is a concurrent web spider that stays within a single domain. It has three internal packages that form a pipeline:

**`internal/reader`** — HTTP fetching. `Reader.Fetch` wraps `http.Client` and returns the response body as a streaming `io.ReadCloser` (caller closes it). It rejects non-2xx responses and non-`text/html` content types, sets a `User-Agent`, and caps the body with `io.LimitReader`. This is the only package that makes network calls.

**`internal/parser`** — HTML link extraction. `ParseLinks` tokenizes HTML using `golang.org/x/net/html`, filters `<a href>` values to same-host links only, strips fragments, rejects relative traversals (`../`) and known binary extensions via the `skipExtensions` denylist (note: it deliberately keeps `.html`/`.php`/etc.), and deduplicates results.

**`internal/crawler`** — Orchestration. `Crawler.Start` runs a single coordinator loop plus one `crawl` goroutine per discovered URL:
- The coordinator owns the `seen` set, the result slice, and an `outstanding` counter — all accessed from one goroutine, so no locking is needed. It receives `Page` results, records them, and enqueues unseen links.
- `crawl` fetches and parses one page and sends exactly one `Page` on the results channel (an empty `Page` on failure, which the coordinator drops). A `sem` buffered-channel semaphore of size `workers` bounds concurrent fetches; the result send selects on `ctx.Done()`.

Concurrency is bounded by the semaphore (the channel buffer is *not* the bound). The coordinator returns as soon as `outstanding` hits zero or `ctx` is cancelled, leaving no goroutines behind. The `Reader` dependency is an interface, mocked in tests with a `stubReader`. Graceful shutdown is wired via `signal.NotifyContext` in `main.go`.

## I/O contract

Results print to **stdout** (`Page <url> has links: [...]`); structured JSON logs go to **stderr**. Keep stdout pipeable — don't log to it.

## Non-goals (intentional)

No robots.txt, depth/page cap, rate limiting, retries, or cross-run state. Don't add these as "fixes" unless asked — single-domain in-memory crawl is the scope.
