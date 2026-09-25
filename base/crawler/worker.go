package crawler

import (
	"context"
	"log/slog"

	"ints-test-assign/base/fetcher"
	"ints-test-assign/base/models"
	"ints-test-assign/base/parser"
)

type Worker struct {
	Fetcher fetcher.Fetcher
	Parser  parser.Parser
	Logger  *slog.Logger
}

func NewWorker(fetcher fetcher.Fetcher, parser parser.Parser, logger *slog.Logger) *Worker {
	return &Worker{
		Fetcher: fetcher,
		Parser:  parser,
		Logger:  logger,
	}
}

func (w *Worker) process(ctx context.Context, task models.Task) models.Result {
	data, err := w.Fetcher.Fetch(ctx, task.URL)
	if err != nil {
		return models.Result{
			Task:  task,
			Error: err,
		}
	}

	page, err := w.Parser.Parse(data)
	if err != nil {
		return models.Result{
			Task:  task,
			Error: err,
		}
	}

	return models.Result{
		Task: task,
		Page: &models.Page{
			Resource: task.URL,
			Title:    page.Title,
			Links:    make([]models.Page, 0),
		},
	}
}

func (w *Worker) Run(ctx context.Context, tasks <-chan models.Task, results chan<- models.Result) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-tasks:
			if !ok {
				return
			}

			result := w.process(ctx, task)

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}

		}
	}
}
