package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "test.log")

	logger, closeLogger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if logger == nil {
		t.Fatal("expected logger, got nil")
	}

	if closeLogger == nil {
		t.Fatal("expected close function, got nil")
	}

	if err := closeLogger(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log file was not created: %v", err)
	}
}

func TestNewLoggerMakesLogs(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "test.log")

	logger, closeLogger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	logger.Info("test message", "url", "https://google.com")

	if err := closeLogger(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "test message") {
		t.Errorf("log does not contain message: %s", content)
	}

	if !strings.Contains(content, "https://google.com") {
		t.Errorf("log does not contain URL: %s", content)
	}
}

func TestNewLoggerInvalidPath(t *testing.T) {
	invalidPath := filepath.Join(t.TempDir(), "nonexistent", "directory", "test.log")

	logger, closeLogger, err := NewLogger(invalidPath)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if logger != nil {
		t.Error("expected nil logger")
	}

	if closeLogger != nil {
		t.Error("expected nil close function")
	}
}

func TestNewLoggerJustWorks(t *testing.T) { // prekols) смысла нет но пусть будет
	logPath := filepath.Join(t.TempDir(), "test.log")

	logger, closeLogger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer closeLogger()

	var _ *slog.Logger = logger
}
