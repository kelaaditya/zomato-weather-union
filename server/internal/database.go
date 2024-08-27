package internal

import (
	"context"

	// external
	"github.com/jackc/pgx/v5/pgxpool"
)

// create a database pool from context and database URL
// and return a pointer to the pool
func (appConfig *AppConfig) DBPoolInitialize(
	context context.Context,
	databaseURL string,
) error {
	// create new pool using the
	// app context and the database URL
	appDBPool, err := pgxpool.New(context, databaseURL)
	if err != nil {
		return err
	}

	// setting the config struct to carry the
	// newly configured database connection pool
	appConfig.DBPool = appDBPool

	// return nil if all okay
	return nil
}

// ping database for connection check
func (appConfig *AppConfig) DBPoolPing(context context.Context) error {
	err := appConfig.DBPool.Ping(context)
	return err
}

// function for closing the connection
func (appConfig *AppConfig) DBPoolClose() {
	appConfig.DBPool.Close()
}
