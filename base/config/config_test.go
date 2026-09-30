package config

import (
	"testing"
	"time"
)

func validConfig() *Config {
	urls := "https://google.com"
	depth := 1
	workers := 10
	timeout := 10 * time.Second
	requestTimeout := 5 * time.Second
	output := "output.json"
	log := "crawler.log"

	return &Config{
		URLs:           urls,
		Depth:          depth,
		Workers:        workers,
		Timeout:        timeout,
		RequestTimeout: requestTimeout,
		Output:         output,
		Log:            log,
	}
}

func TestConfigValidateNormal(t *testing.T) {
	config := validConfig()

	if err := config.validate(); err != nil {
		t.Fatalf("expected config to be valid, got error: %v", err)
	}
}

func TestConfigValidateDepthZero(t *testing.T) {
	config := validConfig()

	depth := 0
	config.Depth = depth

	if err := config.validate(); err != nil {
		t.Fatalf("depth=0 should be valid, got error: %v", err)
	}
}

func TestConfigValidateNegativeDepth(t *testing.T) {
	config := validConfig()

	depth := -1
	config.Depth = depth

	if err := config.validate(); err == nil {
		t.Fatal("expected error for negative depth")
	}
}

func TestConfigValidateWorkersLimit(t *testing.T) {
	tests := []struct {
		name    string
		workers int
		wantErr bool
	}{
		{
			name:    "minimum workers",
			workers: 1,
			wantErr: false,
		},
		{
			name:    "maximum workers",
			workers: 10,
			wantErr: false,
		},
		{
			name:    "too many workers",
			workers: 11,
			wantErr: true,
		},
		{
			name:    "zero workers",
			workers: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()

			workers := tt.workers
			config.Workers = workers

			err := config.validate()

			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected config to be valid, got error: %v", err)
			}
		})
	}
}

func TestConfigValidateEmptyURLs(t *testing.T) {
	config := validConfig()

	urls := ""
	config.URLs = urls

	if err := config.validate(); err == nil {
		t.Fatal("expected error for empty URLs")
	}
}

func TestConfigValidateTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "positive timeout",
			timeout: 10 * time.Second,
			wantErr: false,
		},
		{
			name:    "zero timeout",
			timeout: 0 * time.Second,
			wantErr: true,
		},
		{
			name:    "negative timeout",
			timeout: -1 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()

			timeout := tt.timeout
			config.Timeout = timeout

			err := config.validate()

			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected config to be valid, got error: %v", err)
			}
		})
	}
}

func TestConfigValidateRequestTimeout(t *testing.T) {
	config := validConfig()

	timeout := 0 * time.Second
	config.RequestTimeout = timeout

	if err := config.validate(); err == nil {
		t.Fatal("expected error for zero request timeout")
	}
}

func TestConfigValidateOutput(t *testing.T) {
	config := validConfig()

	output := ""
	config.Output = output

	if err := config.validate(); err == nil {
		t.Fatal("expected error for empty output")
	}
}

func TestConfigValidateLog(t *testing.T) {
	config := validConfig()

	log := ""
	config.Log = log

	if err := config.validate(); err == nil {
		t.Fatalf("expected error for empty if specified log")
	}
}
