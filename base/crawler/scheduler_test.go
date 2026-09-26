package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"ints-test-assign/base/fetcher"
	"ints-test-assign/base/parser"
)

func TestScheduler_Crawl(t *testing.T) {
	var requests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		fmt.Println("REQUEST:", r.URL.Path)

		w.Header().Set("Content-Type", "text/html")

		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, `
				<html>
					<head><title>Test Page</title></head>
					<body>
						<a href="/about">About</a>
						<a href="/contacts">Contacts</a>
						<a href="/about">About again</a>
						<a href="https://google.com">Google</a>
					</body>
				</html>
			`)

		case "/about":
			fmt.Fprint(w, `
				<html>
					<head><title>About</title></head>
					<body>
						<a href="/team">Team</a>
					</body>
				</html>
			`)

		case "/contacts":
			fmt.Fprint(w, `
				<html>
					<head><title>Contacts</title></head>
					<body>
						<a href="/">Home</a>
					</body>
				</html>
			`)

		case "/team":
			fmt.Fprint(w, `
				<html>
					<head><title>Team</title></head>
				</html>
			`)

		default:
			http.NotFound(w, r)
		}
	}))

	defer server.Close()

	ctx := context.Background()

	httpFetcher := fetcher.NewHTTPFetcher(5 * time.Second)
	htmlParser := parser.NewHTMLParser()
	logger := slog.Default()
	worker := NewWorker(httpFetcher, htmlParser, logger)
	scheduler := NewScheduler(3)

	startURL := server.URL + "/"

	result := scheduler.Run(ctx, []string{startURL}, 2, worker)

	if len(result) != 1 {
		t.Fatalf("expected 1 root, got %v", len(result))
	}

	root := result[0]

	if root.Resource != startURL {
		t.Fatalf("expected resource %v, got %v", []byte(startURL), []byte(root.Resource))
	}

	if root.Title != "Test Page" {
		t.Fatalf("expected title 'Test Page', got %q", root.Title)
	}

	if len(root.Links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(root.Links))
	}

	if atomic.LoadInt32(&requests) != 4 {
		t.Fatalf("expected 4 requests, got %d", requests)
	}
}
