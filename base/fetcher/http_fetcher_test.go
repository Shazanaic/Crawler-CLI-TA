package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHttpFetcher_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>Hello</body></html>"))
	}))

	defer server.Close()

	f := NewHTTPFetcher(5 * time.Second)

	body, err := f.Fetch(context.Background(), server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "<html><body>Hello</body></html>"
	if strings.TrimSpace(string(body)) != expected {
		t.Fatalf("expected body %q, got %q", expected, string(body))
	}
}

func TestHttpFetcher_Redirection(t *testing.T) {
	redirected := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !redirected {
			redirected = true
			http.Redirect(w, r, "/new-location", http.StatusFound)
			return
		}

		if r.URL.Path == "/page" {
			redirected = true

			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><body>Redirected</body></html>"))
			return
		}
	}))
	defer server.Close()

	f := NewHTTPFetcher(5 * time.Second)
	_, err := f.Fetch(context.Background(), server.URL)

	if err == nil {
		t.Fatalf("expected error due to redirection, got nil")
	}
}

func TestHttpFetcher_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	f := NewHTTPFetcher(5 * time.Second)
	_, err := f.Fetch(context.Background(), server.URL)
	if err == nil {
		t.Fatalf("expected error due to 404, got nil")
	}
}

func TestHttpFetcher_NonHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Hello"}`))
	}))

	defer server.Close()

	f := NewHTTPFetcher(5 * time.Second)
	_, err := f.Fetch(context.Background(), server.URL)
	if err == nil {
		t.Fatalf("expected error due to non-HTML content, got nil")
	}
}

func TestHttpFetcher_RequestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>timeout page</body></html>"))
	}))

	defer server.Close()

	f := NewHTTPFetcher(1 * time.Second)

	start := time.Now()

	_, err := f.Fetch(context.Background(), server.URL)

	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected error due to request timeout, got nil")
	}

	if elapsed > 2*time.Second {
		t.Fatalf("timeout is longer than expected: %v", elapsed)
	}
}

func TestHttpFetcher_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>cancelled page</body></html>"))
	}))

	defer server.Close()

	f := NewHTTPFetcher(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	cancel()

	_, err := f.Fetch(ctx, server.URL)
	if err == nil {
		t.Fatalf("expected error due to context cancellation, got nil")
	}
}
