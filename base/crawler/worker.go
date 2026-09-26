package crawler

import (
	"context"
	"log/slog"

	"Crawler-CLI-TA/base/fetcher"
	"Crawler-CLI-TA/base/models"
	"Crawler-CLI-TA/base/parser"
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
	resp, err := w.Fetcher.Fetch(ctx, task.URL)

	if w.Logger != nil && resp.StatusCode != 0 {
		w.Logger.Info(
			"http request",
			"url", task.URL,
			"status_code", resp.StatusCode,
			"status", resp.Status,
		)
	}

	if err != nil {
		if w.Logger != nil {
			w.Logger.Error(
				"failed to fetch page",
				"url", task.URL,
				"error", err,
			)
		}

		return models.Result{
			Task:  task,
			Error: err,
		}
	}

	page, err := w.Parser.Parse(resp.Body)
	if err != nil {
		if w.Logger != nil {
			w.Logger.Error(
				"failed to parse page",
				"url", task.URL,
				"error", err,
			)
		}

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
			Links:    make([]*models.Page, 0),
		},
		Links: page.URLs,
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
