package fetcher

import (
	"context"
)

type Response struct {
	Body       []byte
	StatusCode int
	Status     string
}

type Fetcher interface {
	Fetch(ctx context.Context, url string) (Response, error)
}
