package internal

import (
	"log/slog"
	"os"
)

// create logger at the app level
func (appConfig *AppConfig) LoggerInitialize() {
	// logger configuration options
	loggerOptions := slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	// logger handler
	loggerHandler := slog.NewTextHandler(os.Stdout, &loggerOptions)

	// create new logger from handler
	logger := slog.New(loggerHandler)

	// setting the config struct to carry the
	// newly configured logger
	appConfig.Logger = logger
}
