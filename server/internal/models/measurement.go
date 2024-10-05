package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// model struct for measurements
type MeasurementModel struct {
	DB *pgxpool.Pool
}

// structure of measurement data needed for wet bulb calculations
type MeasurementTemperature struct {
	MeasurementIDWeatherUnion   uuid.UUID `json:"measurement_id_weather_union"`
	MeasurementIDOpenWeatherMap uuid.UUID `json:"measurement_id_open_weather_map"`
	WeatherStationID            uuid.UUID `json:"weather_station_id"`
	RunID                       uuid.UUID `json:"run_id"`
	Temperature                 float64   `json:"temperature"`
	Humidity                    float64   `json:"humidity"`
	Pressure                    float64   `json:"pressure"`
}

// function to save the measurement run ID
func (model MeasurementModel) SaveMeasurementRun(
	ctx context.Context,
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

	// create a 5 second timeout context
	ctxWT, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancellation of the timeout
	defer cancel()

	// executing the query string with the named arguments
	_, err := model.DB.Exec(ctxWT, queryString, queryArguments)
	if err != nil {
		return fmt.Errorf(
			"error in inserting measurement run data into postgresql: %w",
			err,
		)
	}

	return nil
}

// get all the unprocessed measurement values
// from both weather union and open weather map
func (model MeasurementModel) GetUnprocessedDataForCalculationsTemperature(
	ctx context.Context,
) ([]MeasurementTemperature, error) {
	// postgresql query string
	var queryString string = `
	SELECT
		mwu.measurement_id AS measurement_id_weather_union,
		mwu.weather_station_id,
		mwu.run_id,
		mwu.temperature,
		mwu.humidity,
		mowm.measurement_id AS measurement_id_open_weather_map,
		mowm.pressure
	FROM measurements_weather_union mwu
	JOIN measurements_open_weather_map mowm
	ON
		mwu.weather_station_id = mowm.weather_station_id AND
		mwu.run_id = mowm.run_id
	WHERE
		mwu.is_processed_for_calculation_temperature = FALSE AND
		mowm.is_processed_for_calculation_temperature = FALSE AND
		mwu.temperature IS NOT NULL AND
		mwu.humidity IS NOT NULL AND
		mowm.pressure IS NOT NULL;
	`

	// create a 5 second timeout context
	ctxWT, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancellation of the timeout
	defer cancel()

	// prepare the query
	rows, err := model.DB.Query(ctxWT, queryString)
	if err != nil {
		return nil, err
	}

	// run the query and collect rows
	unprocessedSlice, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[MeasurementTemperature],
	)
	if err != nil {
		return nil, err
	}

	// return slice of unprocessed data for wet bulb calculations
	return unprocessedSlice, nil
}
