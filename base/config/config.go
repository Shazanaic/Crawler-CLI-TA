package config

import (
	"errors"
	"flag"
	"strings"
	"time"
)

type Config struct {
	URLs           string
	Depth          int
	Workers        int
	Timeout        time.Duration
	RequestTimeout time.Duration
	Output         string
	Log            string
}

func Parse() (*Config, error) {
	var urls string
	var depth int
	var workers int
	var timeout time.Duration
	var requestTimeout time.Duration
	var output string
	var log string

	flag.StringVar(&urls, "urls", "", "Comma-separated list of URLs to crawl")
	flag.IntVar(&depth, "depth", 1, "Depth of crawling")
	flag.IntVar(&workers, "workers", 10, "Maximum number of workers")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "Timeout for the entire crawling process")
	flag.DurationVar(&requestTimeout, "request-timeout", 5*time.Second, "Timeout for individual HTTP requests")
	flag.StringVar(&output, "output", "output.txt", "Output file for the results")
	flag.StringVar(&log, "log", "logs.log", "Log file for logging")
	flag.Parse()

	config := &Config{
		URLs:           urls,
		Depth:          depth,
		Workers:        workers,
		Timeout:        timeout,
		RequestTimeout: requestTimeout,
		Output:         output,
		Log:            log,
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Config) validate() error {
	if c.URLs == "" {
		return errors.New("urls cannot be empty")
	}
	if c.Depth < 1 {
		return errors.New("depth must be at least 1")
	}
	if c.Workers < 1 {
		return errors.New("workers must be at least 1")
	}
	if c.Workers > 10 {
		return errors.New("workers cant be greater than 10")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}
	if c.RequestTimeout <= 0 {
		return errors.New("request timeout must be greater than 0")
	}
	if c.Output == "" {
		return errors.New("output file cannot be empty")
	}
	if c.Log != "" && strings.TrimSpace(c.Log) == "" {
		return errors.New("log file cannot be empty if specified")
	}
	return nil
}
