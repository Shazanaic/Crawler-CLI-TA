package crawler

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"Crawler-CLI-TA/base/models"
)

const maxWorkers = 10

type Scheduler struct {
	workers int
}

func NewScheduler(workers int) *Scheduler {
	if workers <= 0 {
		workers = 1
	}
	if workers > maxWorkers {
		workers = maxWorkers
	}
	return &Scheduler{
		workers: workers,
	}
}

func (sch *Scheduler) Run(ctx context.Context, startURLs []string, maxDepth int, worker *Worker) []models.Page {
	tasks := make(chan models.Task)
	results := make(chan models.Result)

	var wg sync.WaitGroup

	for i := 0; i < sch.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker.Run(ctx, tasks, results)
		}()
	}

	visited := make(map[string]bool)
	roots := make([]*models.Page, 0)
	pending := make([]models.Task, 0)

	for _, url := range startURLs {
		rootUrl, err := getDomain(url)
		if err != nil {
			continue
		}

		if visited[url] {
			continue
		}

		visited[url] = true
		pending = append(pending, models.Task{
			URL:     url,
			Depth:   0,
			RootURL: rootUrl,
			Parent:  nil,
		})
	}

	active := 0

schLoop:
	for len(pending) > 0 || active > 0 {
		if ctx.Err() != nil {
			break
		}

		if len(pending) > 0 && active < sch.workers {
			task := pending[0]
			pending[0] = models.Task{}
			pending = pending[1:]

			select {
			case tasks <- task:
				active++

			case <-ctx.Done():
				break schLoop
			}

			continue
		}

		select {
		case result := <-results:
			active--
			sch.handleResult(result, maxDepth, visited, &pending, &roots)

		case <-ctx.Done():
			break schLoop
		}
	}

	close(tasks)
	wg.Wait()
	return convertRootsToPages(roots)
}

func (sch *Scheduler) handleResult(result models.Result, maxDepth int, visited map[string]bool, pending *[]models.Task, roots *[]*models.Page) {
	if result.Error != nil {
		return
	}

	if result.Page == nil {
		return
	}

	page := result.Page
	if result.Task.Parent == nil {
		*roots = append(*roots, page)
	} else {
		result.Task.Parent.Links = append(result.Task.Parent.Links, page)
	}

	if result.Task.Depth >= maxDepth {
		return
	}

	for _, link := range sch.getAllowedLinks(result.Task.URL, result.Task.RootURL, result.Links) {
		if visited[link] {
			continue
		}

		visited[link] = true
		*pending = append(*pending, models.Task{
			URL:     link,
			Depth:   result.Task.Depth + 1,
			RootURL: result.Task.RootURL,
			Parent:  page,
		})
	}
}

func (sch *Scheduler) getAllowedLinks(currentURL, rootURL string, links []string) []string {
	allowedLinks := make([]string, 0)

	seen := make(map[string]struct{}) // чтоб не спамить одинаковыми ссылками в слайсыч из-за ResolveReference,

	baseURL, err := url.Parse(currentURL)
	if err != nil {
		return allowedLinks
	}

	for _, link := range links {
		linkURL, err := url.Parse(link)

		if err != nil {
			continue
		}

		resolvedURL := baseURL.ResolveReference(linkURL)

		if resolvedURL.Scheme != "http" && resolvedURL.Scheme != "https" {
			continue
		}

		if resolvedURL.Hostname() != rootURL {
			continue
		}

		resolvedURLStr := resolvedURL.String()

		if _, exists := seen[resolvedURLStr]; exists {
			continue
		}

		seen[resolvedURLStr] = struct{}{}
		allowedLinks = append(allowedLinks, resolvedURLStr)
	}

	return allowedLinks
}

func getDomain(urlRaw string) (string, error) {
	parsedURL, err := url.Parse(urlRaw)

	if err != nil {
		return "", err
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", errors.New("unsupported scheme")
	}

	if parsedURL.Hostname() == "" {
		return "", errors.New("no hostname")
	}

	return parsedURL.Hostname(), nil
}

func convertRootsToPages(roots []*models.Page) []models.Page {
	result := make([]models.Page, 0, len(roots))

	for _, root := range roots {
		if root != nil {
			result = append(result, *root)
		}
	}

	return result
}
