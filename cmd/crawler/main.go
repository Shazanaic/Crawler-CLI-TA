package main

import (
	"Crawler-CLI-TA/base/config"
	"Crawler-CLI-TA/base/crawler"
	"Crawler-CLI-TA/base/logging"
	"Crawler-CLI-TA/base/output"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func splitURLs(value string) []string {
	parts := strings.Split(value, ",")

	urls := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part != "" {
			urls = append(urls, part)
		}
	}

	return urls
}

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("error parsing config:%v", err)
	}

	fmt.Println("Config:", cfg)

	logger, closeLogger, err := logging.NewLogger(cfg.Log)
	if err != nil {
		log.Fatalf("logger error:%v", err)
	}
	defer closeLogger()

	jsonWriter := output.NewJSONWriter()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	crawler := crawler.NewCrawler(cfg.Workers, cfg.RequestTimeout, logger)

	startURLs := splitURLs(cfg.URLs)

	results := crawler.Run(ctx, startURLs, cfg.Depth)

	if err := jsonWriter.Write(results, cfg.Output); err != nil {
		logger.Error("failed to json-write results", "error", err)

		log.Fatalf("output error:%v", err)
	}

	if ctx.Err() == context.Canceled {
		fmt.Println("stopped by user")
		return
	}

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("stopped by global timeout")
		return
	}
}
