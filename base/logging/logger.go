package logging

import (
	"log/slog"
	"os"
)

func NewLogger(path string) (*slog.Logger, func() error, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, err
	}

	logger := slog.New(slog.NewTextHandler(file, nil))

	closeFunc := func() error {
		return file.Close()
	}

	return logger, closeFunc, nil
}
