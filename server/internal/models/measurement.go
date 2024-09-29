package models

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kelaaditya/zomato-weather-union/server/internal"
)

func SaveMeasurementRun(
	ctx context.Context,
	appConfig *internal.AppConfig,
	runID uuid.UUID,
) error {
	// postgresql query string
	var queryString string = `
	INSERT INTO measurement_runs(run_id)
	VALUES (@runID);
	`

	// named arguments for building the query string
	var queryArguments pgx.NamedArgs = pgx.NamedArgs{
		"runID": runID,
	}

	// executing the query string with the named arguments
	_, err := appConfig.DBPool.Exec(ctx, queryString, queryArguments)
	if err != nil {
		return fmt.Errorf(
			"error in inserting measurement run data into postgresql: %w",
			err,
		)
	}

	return nil
}
