package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"Crawler-CLI-TA/base/fetcher"
	"Crawler-CLI-TA/base/parser"
)

type testFetcher struct{}

func (f *testFetcher) Fetch(ctx context.Context, url string) (fetcher.Response, error) {
	return fetcher.Response{
		Body:       []byte(url),
		StatusCode: 200,
		Status:     "ok",
	}, nil
}

type testParser struct{}

func (p *testParser) Parse(html []byte) (parser.ParseResult, error) {
	url := string(html)

	switch url {
	case "http://example.com/":
		return parser.ParseResult{
			Title: "Home",
			URLs:  []string{"/about"},
		}, nil

	case "http://example.com/about":
		return parser.ParseResult{
			Title: "About",
			URLs:  []string{"/team"},
		}, nil

	case "http://example.com/team":
		return parser.ParseResult{
			Title: "Team",
			URLs:  []string{},
		}, nil
	}

	return parser.ParseResult{}, nil
}

func TestScheduler_BuildsPageTree(t *testing.T) {
	worker := NewWorker(&testFetcher{}, &testParser{}, nil)

	scheduler := NewScheduler(1)

	pages := scheduler.Run(context.Background(), []string{"http://example.com/"}, 2, worker)

	if len(pages) != 1 {
		t.Fatalf("expected 1 root page, got %d", len(pages))
	}

	root := pages[0]

	if root.Resource != "http://example.com/" {
		t.Fatalf("unexpected root resource: %s", root.Resource)
	}

	if len(root.Links) != 1 {
		t.Fatalf(
			"expected root to have 1 child, got %d",
			len(root.Links),
		)
	}

	about := root.Links[0]

	if about.Resource != "http://example.com/about" {
		t.Fatalf(
			"expected about page, got %s",
			about.Resource,
		)
	}

	if len(about.Links) != 1 {
		t.Fatalf(
			"expected about to have 1 child, got %d",
			len(about.Links),
		)
	}

	team := about.Links[0]

	if team.Resource != "http://example.com/team" {
		t.Fatalf("expected team page, got %s", team.Resource)
	}
}

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
	worker := NewWorker(httpFetcher, htmlParser, nil)
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

	httpFetcher := fetcher.NewHTTPFetcher(5 * time.Second)
	htmlParser := parser.NewHTMLParser()
	worker := NewWorker(httpFetcher, htmlParser, nil)
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

	httpFetcher := fetcher.NewHTTPFetcher(5 * time.Second)
	htmlParser := parser.NewHTMLParser()
	worker := NewWorker(httpFetcher, htmlParser, nil)
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

	httpFetcher := fetcher.NewHTTPFetcher(5 * time.Second)
	htmlParser := parser.NewHTMLParser()
	worker := NewWorker(httpFetcher, htmlParser, nil)
	scheduler := NewScheduler(10)

	scheduler.Run(ctx, startURLs, 0, worker)

	if maxActive > 10 {
		t.Fatalf("exeeded max num of active workers:%v", maxActive)
	}
}
