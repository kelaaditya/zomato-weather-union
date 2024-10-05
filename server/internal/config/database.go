package config

import (
	"context"
	"time"

	// external
	"github.com/jackc/pgx/v5/pgxpool"
)

// create a database pool from context and database URL
// and return a pointer to the pool
func (config *Config) InitializeDatabase(
	ctx context.Context,
	databaseURL string,
) error {
	// create new pool using the database URL
	DB, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return err
	}

	// create a 5 second context with timeout for the ping check
	ctxWT, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer the cancel of the context
	defer cancel()
	// ping check
	err = DB.Ping(ctxWT)
	if err != nil {
		return err
	}

	// newly configured database connection pool
	config.DB = DB

	// return nil if all okay
	return nil
}
