package internal

import (
	"context"

	// external
	"github.com/jackc/pgx/v5/pgxpool"
)

// create a database pool from context and database URL
// and return a pointer to the pool
func CreateDBPool(context *context.Context, databaseURL string) (*pgxpool.Pool, error) {
	// create new pool using the
	// app context and the database URL
	appDBPool, err := pgxpool.New(*context, databaseURL)
	if err != nil {
		return nil, err
	}

	// return pointer to created database pool
	return appDBPool, nil
}

// ping database for connection check
func AppDBPoolPing(context *context.Context, databasePool *pgxpool.Pool) error {
	err := databasePool.Ping(*context)
	return err
}

// function for closing the connection
func AppDatabasePoolClose(databasePool *pgxpool.Pool) {
	databasePool.Close()
}
