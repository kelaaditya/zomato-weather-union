package internal

import (
	"log/slog"
	"os"
)

// create logger at the app level
func CreateLogger() *slog.Logger {
	// logger configuration options
	loggerOptions := slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	// logger handler
	loggerHandler := slog.NewTextHandler(os.Stdout, &loggerOptions)

	// create new logger from handler
	logger := slog.New(loggerHandler)

	return logger
}
