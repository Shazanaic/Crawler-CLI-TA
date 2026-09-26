package crawler

import (
	"context"
	"log/slog"
	"time"

	"Crawler-CLI-TA/base/fetcher"
	"Crawler-CLI-TA/base/models"
	"Crawler-CLI-TA/base/parser"
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
