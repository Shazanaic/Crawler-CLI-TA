package crawler

import (
	"context"
	"log/slog"
	"time"

	"ints-test-assign/base/fetcher"
	"ints-test-assign/base/models"
	"ints-test-assign/base/parser"
)

type Crawler struct {
	Scheduler *Scheduler
	Worker    *Worker
	Logger    *slog.Logger
}

func NewCrawler(workers int, requestTimeuot time.Duration, logger *slog.Logger) *Crawler {
	httpFetcher := fetcher.NewHTTPFetcher(requestTimeuot)
	htmlParser := parser.NewHTMLParser()

	worker := NewWorker(httpFetcher, htmlParser, logger)
	scheduler := NewScheduler(workers)

	return &Crawler{
		Scheduler: scheduler,
		Worker:    worker,
		Logger:    logger,
	}
}

func (cr *Crawler) Run(ctx context.Context, startURLs []string, maxDepth int) []models.Page {
	return cr.Scheduler.Run(ctx, startURLs, maxDepth, cr.Worker)
}
