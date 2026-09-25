package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPFetcher struct {
	Client          *http.Client
	request_timeout time.Duration
}

func NewHTTPFetcher(request_timeout time.Duration) *HTTPFetcher {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return &HTTPFetcher{
		Client:          client,
		request_timeout: request_timeout,
	}
}

func (hf *HTTPFetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	rctx, cancel := context.WithTimeout(ctx, hf.request_timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(rctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := hf.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return nil, fmt.Errorf("redirection %s", resp.Status)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	resp.Body, err = CheckHTML(resp)

	if err != nil {
		return nil, fmt.Errorf("unsupported content type: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

// проверка на html, учитывая возможность отсутствия хедеров типа контента
// подглядывает через 512 первых байт(или меньше) и определяет тип контента страницы
func CheckHTML(resp *http.Response) (io.ReadCloser, error) {
	content := resp.Header.Get("Content-Type")
	if resp.Header.Get("Content-Type") != "" {
		if strings.HasPrefix(strings.ToLower(content), "text/html") {
			return resp.Body, nil
		}

		resp.Body.Close()
		return nil, fmt.Errorf("unsupported content type: %s", content)
	}

	buf := make([]byte, 512)
	n, err := io.ReadFull(resp.Body, buf)

	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	readBuf := buf[:n]

	contentType := http.DetectContentType(readBuf)

	if !strings.HasPrefix(strings.ToLower(contentType), "text/html") {
		resp.Body.Close()
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	reader := io.MultiReader(
		strings.NewReader(string(readBuf)),
		resp.Body,
	)

	return struct {
		io.Reader
		io.Closer
	}{reader, resp.Body}, nil
}
