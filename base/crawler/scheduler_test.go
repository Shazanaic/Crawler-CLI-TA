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

	startURL := server.URL + "/" //иначе путает и считает ссылку на старт новой ссылкой

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

func TestScheduler_MaxDepth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, `
				<title>Root</title>
				<a href="/level1">Level 1</a>
			`)

		case "/level1":
			fmt.Fprint(w, `
				<title>Level 1</title>
				<a href="/level2">Level 2</a>
			`)

		case "/level2":
			fmt.Fprint(w, `
				<title>Level 2</title>
				<a href="/level3">Level 3</a>
			`)

		case "/level3":
			fmt.Fprint(w, `
				<title>Level 3</title>
			`)
		}
	}))

	defer server.Close()
	startURL := server.URL + "/"

	ctx := context.Background()
	logger := slog.Default()
	httpFetcher := fetcher.NewHTTPFetcher(5 * time.Second)
	htmlParser := parser.NewHTMLParser()
	worker := NewWorker(httpFetcher, htmlParser, logger)
	scheduler := NewScheduler(2)

	result := scheduler.Run(ctx, []string{startURL}, 1, worker)

	if len(result) != 1 {
		t.Fatalf("expected 1 root, got %v", len(result))
	}

	if len(result[0].Links) != 1 {
		t.Fatalf("expected 1 child page, got %v", len(result[0].Links))
	}

	if result[0].Links[0].Resource != startURL+"level1" {
		t.Fatalf("unexpected child page URL: %s", result[0].Links[0].Resource)
	}

	if len(result[0].Links[0].Links) != 0 {
		t.Fatal("max depth exeeded")
	}
}

func TestScheduler_NoCycles(t *testing.T) {
	var requests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)

		w.Header().Set("Content-Type", "text/html")

		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, `
				<title>Root</title>
				<a href="/page1">Page 1</a>
			`)

		case "/page1":
			fmt.Fprint(w, `
				<title>Page 1</title>
				<a href="/">Root</a>
				<a href="/page2">Page 2</a>
			`)

		case "/page2":
			fmt.Fprint(w, `
				<title>Page 2</title>
				<a href="/page1">Page 1</a>
			`)
		}
	}))

	defer server.Close()
	startURL := server.URL + "/"

	ctx := context.Background()
	logger := slog.Default()

	httpFetcher := fetcher.NewHTTPFetcher(5 * time.Second)
	htmlParser := parser.NewHTMLParser()
	worker := NewWorker(httpFetcher, htmlParser, logger)
	scheduler := NewScheduler(3)

	result := scheduler.Run(ctx, []string{startURL}, 10, worker)

	if len(result) != 1 {
		t.Fatalf("expected 1 root, got %v", len(result))
	}

	if atomic.LoadInt32(&requests) != 3 {
		t.Fatalf("expected 3 requests, got %v", requests)
	}
}

func TestScheduler_MaxWorkers(t *testing.T) {
	var active int32
	var maxActive int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&active, 1)

		for {
			old := atomic.LoadInt32(&maxActive)

			if current <= old {
				break
			}

			if atomic.CompareAndSwapInt32(&maxActive, old, current) {
				break
			}
		}

		defer atomic.AddInt32(&active, -1)

		w.Header().Set("Content-Type", "text/html")

		fmt.Fprint(w, `<title>Test Page</title>`)
	}))

	defer server.Close()
	startURLs := make([]string, 0, 20)

	for i := 0; i < 20; i++ {
		startURLs = append(startURLs, fmt.Sprintf("%s/page%d", server.URL, i))
	}
	ctx := context.Background()
	logger := slog.Default()

	httpFetcher := fetcher.NewHTTPFetcher(5 * time.Second)
	htmlParser := parser.NewHTMLParser()
	worker := NewWorker(httpFetcher, htmlParser, logger)
	scheduler := NewScheduler(10)

	scheduler.Run(ctx, startURLs, 0, worker)

	if maxActive > 10 {
		t.Fatalf("exeeded max num of active workers:%v", maxActive)
	}
}
