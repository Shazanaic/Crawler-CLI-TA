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
	Timeout        time.Duration
	RequestTimeout time.Duration
	Output         string
	Log            string
}

func Parse() (Config, error) {
	var urls string
	var depth int
	var timeout time.Duration
	var requestTimeout time.Duration
	var output string
	var log string

	flag.StringVar(&urls, "urls", "", "Comma-separated list of URLs to crawl")
	flag.IntVar(&depth, "depth", 1, "Depth of crawling")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "Timeout for the entire crawling process")
	flag.DurationVar(&requestTimeout, "request-timeout", 5*time.Second, "Timeout for individual HTTP requests")
	flag.StringVar(&output, "output", "output.txt", "Output file for the results")
	flag.StringVar(&log, "log", "", "Log file for logging")
	flag.Parse()

	config := Config{
		URLs:           urls,
		Depth:          depth,
		Timeout:        timeout,
		RequestTimeout: requestTimeout,
		Output:         output,
		Log:            log,
	}
	if err := config.validate(); err != nil {
		return Config{}, err
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
