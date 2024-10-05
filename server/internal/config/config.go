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
	config.InitializeLogger()

	// initialize environment
	err := config.InitializeEnvironment()
	if err != nil {
		return err
	}

	// get database URL from environment
	var databaseURL string = config.Environment.DatabaseURL
	// initialize database
	err = config.InitializeDatabase(ctx, databaseURL)
	if err != nil {
		return err
	}

	// return nil if all okay
	return nil
}
