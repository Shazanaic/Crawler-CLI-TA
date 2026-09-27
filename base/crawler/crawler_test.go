package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"Crawler-CLI-TA/base/fetcher"
	"Crawler-CLI-TA/base/parser"
)

func TestCrawler_Run(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, `
				<html>
					<head>
						<title>Test Page</title>
					</head>
					<body>
						<a href="/about">About</a>
					</body>
				</html>
			`)

		case "/about":
			fmt.Fprint(w, `
				<html>
					<head>
						<title>About</title>
					</head>
				</html>
			`)

		default:
			http.NotFound(w, r)
		}
	}))

	defer server.Close()

	crawler := NewCrawler(2, 5*time.Second, slog.Default())

	ctx := context.Background()
	startURL := server.URL + "/"

	result := crawler.Run(ctx, []string{startURL}, 1)

	if len(result) != 1 {
		t.Fatalf("expected 1 root, got %d", len(result))
	}

	if result[0].Title != "Test Page" {
		t.Fatalf("expected title 'Test Page', got %v", result[0].Title)
	}

	if len(result[0].Links) != 1 {
		t.Fatalf("expected 1 child page, got %v", len(result[0].Links))
	}

	if result[0].Links[0].Title != "About" {
		t.Fatalf("expected child title 'About', got %v", result[0].Links[0].Title)
	}
}

func TestCrawler_IntegrationTest(t *testing.T) {
	var (
		mu       sync.Mutex
		requests = make(map[string]int)
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests[r.URL.Path]++
		mu.Unlock()

		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `
				<html>
					<head>
						<title>Test Page</title>
					</head>
					<body>
						<a href="/about">About</a>
						<a href="/missing">Missing</a>
						<a href="/image">Image</a>
					</body>
				</html>
			`)

		case "/about":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `
				<html>
					<head>
						<title>About</title>
					</head>
					<body>
						<a href="/team">Team</a>
					</body>
				</html>
			`)

		case "/team":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `
				<html>
					<head>
						<title>Team</title>
					</head>
					<body></body>
				</html>
			`)

		case "/missing":
			http.Error(w, "not found", http.StatusNotFound)

		case "/image":
			w.Header().Set("Content-Type", "image/jpeg")
			fmt.Fprint(w, "fake image")

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	reqTimeout := 2 * time.Second
	httpFetcher := fetcher.NewHTTPFetcher(reqTimeout)
	htmlParser := parser.NewHTMLParser()
	worker := NewWorker(httpFetcher, htmlParser, nil)

	scheduler := NewScheduler(2)

	ctx := context.Background()

	pages := scheduler.Run(ctx, []string{server.URL + "/"}, 2, worker)

	if len(pages) != 1 {
		t.Fatalf("expected 1 root page, got %v", len(pages))
	}

	root := pages[0]

	if root.Resource != server.URL+"/" {
		t.Fatalf(
			"expected root resource %q, got %q", server.URL+"/", root.Resource)
	}

	if root.Title != "Test Page" {
		t.Fatalf("expected root title %q, got %q", "Test Page", root.Title)
	}

	if len(root.Links) != 1 {
		t.Fatalf("expected root to have 1 child, got %v", len(root.Links))
	}

	about := root.Links[0]

	if about.Resource != server.URL+"/about" {
		t.Fatalf("expected about resource %q, got %q", server.URL+"/about", about.Resource)
	}

	if about.Title != "About" {
		t.Fatalf("expected about title %q, got %q", "About", about.Title)
	}

	if len(about.Links) != 1 {
		t.Fatalf("expected about to have 1 child, got %v", len(about.Links))
	}

	team := about.Links[0]

	if team.Resource != server.URL+"/team" {
		t.Fatalf("expected team resource %q, got %q", server.URL+"/team", team.Resource)
	}

	if team.Title != "Team" {
		t.Fatalf("expected team title %q, got %q", "Team", team.Title)
	}

	if len(team.Links) != 0 {
		t.Fatalf("expected team to have no children, got %v", len(team.Links))
	}

	mu.Lock()
	defer mu.Unlock()

	if requests["/"] != 1 {
		t.Errorf("expected / to be requested once, got %v", requests["/"])
	}

	if requests["/about"] != 1 {
		t.Errorf("expected /about to be requested once, got %v", requests["/about"])
	}

	if requests["/team"] != 1 {
		t.Errorf("expected /team to be requested once, got %v", requests["/team"])
	}

	if requests["/missing"] != 1 {
		t.Errorf("expected /missing to be requested once, got %v", requests["/missing"])
	}

	if requests["/image"] != 1 {
		t.Errorf("expected /image to be requested once, got %v", requests["/image"])
	}
}
