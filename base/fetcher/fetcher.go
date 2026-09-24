package fetcher

import (
	"context"
)

type Fetcher interface {
	Fetch(ctx context.Context, url string) (string, error)
}
