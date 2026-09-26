package crawler

import (
	"context"

	"ints-test-assign/base/models"
)

type Crawler struct {
	Scheduler *Scheduler
	Worker    *Worker
}

func NewCrawler(Scheduler *Scheduler, Worker *Worker) *Crawler {
	return &Crawler{
		Scheduler: Scheduler,
		Worker:    Worker,
	}
}

func (cr *Crawler) Run(ctx context.Context, startURLs []string, maxDepth int) []models.Page {
	return cr.Scheduler.Run(ctx, startURLs, maxDepth, cr.Worker)
}
