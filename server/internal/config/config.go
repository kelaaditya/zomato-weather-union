package config

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DB          *pgxpool.Pool
	Environment *Environment
	Logger      *slog.Logger
}

func (config *Config) New(ctx context.Context) error {
	// initialize logger
	config.initializeLogger()

	// initialize environment
	err := config.initializeEnvironment()
	if err != nil {
		return err
	}

	// get database URL from environment
	var databaseURL string = config.Environment.DatabaseURL
	// initialize database
	err = config.initializeDatabase(ctx, databaseURL)
	if err != nil {
		return err
	}

	// return nil if all okay
	return nil
}
