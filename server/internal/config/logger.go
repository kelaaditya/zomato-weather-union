package config

import (
	"log/slog"
	"os"
)

// create logger at the app level
func (config *Config) initializeLogger() {
	// logger configuration options
	loggerOptions := slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	// logger handler
	newTextHandler := slog.NewTextHandler(os.Stdout, &loggerOptions)

	// create new logger from handler
	newLogger := slog.New(newTextHandler)

	// set newly created logger on the config struct
	config.Logger = newLogger
}
