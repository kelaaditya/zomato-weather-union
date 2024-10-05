package config

import (
	"log/slog"
	"os"
)

// create logger at the app level
func (config *Config) InitializeLogger() {
	// logger configuration options
	loggerOptions := slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	// logger handler
	loggerHandler := slog.NewTextHandler(os.Stdout, &loggerOptions)

	// create new logger from handler
	logger := slog.New(loggerHandler)

	// set newly created logger on the config struct
	config.Logger = logger
}
