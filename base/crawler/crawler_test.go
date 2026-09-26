package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
