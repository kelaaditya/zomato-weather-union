package models

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kelaaditya/zomato-weather-union/server/internal"
	"golang.org/x/sync/errgroup"
)

// structure of measurement data needed for wet bulb calculations
type MeasurementUnprocessedWetBulb struct {
	MeasurementIDWeatherUnion   uuid.UUID `json:"measurement_id_weather_union"`
	MeasurementIDOpenWeatherMap uuid.UUID `json:"measurement_id_open_weather_map"`
	WeatherStationID            uuid.UUID `json:"weather_station_id"`
	RunID                       uuid.UUID `json:"run_id"`
	Temperature                 float64   `json:"temperature"`
	Humidity                    float64   `json:"humidity"`
	Pressure                    float64   `json:"pressure"`
}

func GetUnprocessedDataForWetBulbCalculations(
	ctx context.Context,
	appConfig *internal.AppConfig,
) ([]MeasurementUnprocessedWetBulb, error) {
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
		mwu.temperature IS NOT NULL AND
		mwu.humidity IS NOT NULL AND
		mowm.pressure IS NOT NULL AND
		mwu.is_processed_for_wet_bulb_calculation = FALSE AND
		mowm.is_processed_for_wet_bulb_calculation = FALSE;
	`

	// prepare the query
	rows, err := appConfig.DBPool.Query(ctx, queryString)
	if err != nil {
		return nil, err
	}

	// run the query and collect rows
	unprocessedSlice, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[MeasurementUnprocessedWetBulb],
	)
	if err != nil {
		return nil, err
	}

	// return slice of unprocessed data for wet bulb calculations
	return unprocessedSlice, nil
}

// get all calculations across all weather stations for both weather union
// and open weather map
// this is 'one' run of the server fetcher
func GetAndSaveMeasurementsFromAllWeatherStationsInOneRun(
	ctx context.Context,
	appConfig *internal.AppConfig,
) error {
	// initialize runID as UUID for this run
	runID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	// save measurement run ID
	err = SaveMeasurementRun(ctx, appConfig, runID)
	if err != nil {
		return err
	}

	// get data of all weather stations
	stations, err := GetWeatherStationDataFromWeatherUnion(ctx, appConfig)
	if err != nil {
		return err
	}

	// create a wait group
	var wgCalculations errgroup.Group

	// iterate over stations
	for _, station := range stations {
		wgCalculations.Go(func() error {
			// get measurement from weather union
			err := GetAndSaveMeasurementFromWeatherUnion(
				ctx,
				appConfig,
				runID,
				station,
			)
			if err != nil {
				return err
			}

			// get measurement from open weather map
			err = GetAndSaveMeasurementFromOpenWeatherMap(
				ctx,
				appConfig,
				runID,
				station,
			)
			if err != nil {
				return err
			}

			return nil
		})
	}

	// wait until all goroutines are completed
	// return first non-nil error
	if err := wgCalculations.Wait(); err != nil {
		return err
	}

	return nil
}

// get one measurement from weather union
func GetAndSaveMeasurementFromWeatherUnion(
	ctx context.Context,
	appConfig *internal.AppConfig,
	runID uuid.UUID,
	station WeatherUnionDataStation,
) error {
	// initialize mesurementID as UUID for weather union
	measurementIDWeatherUnion, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	// get data from weather union
	dataWeatherUnion, err := GetWeatherDataFromWeatherUnionLocality(
		appConfig,
		station.LocalityID,
	)
	if err != nil {
		return err
	}

	// save data from weather union
	err = SaveWeatherDataFromWeatherUnion(
		ctx,
		appConfig,
		measurementIDWeatherUnion,
		station.WeatherStationID,
		runID,
		&dataWeatherUnion,
	)
	if err != nil {
		fmt.Println("Error here 2!")
		return err
	}

	// return nil if all okay
	return nil
}

// get one measurement from open weather map
func GetAndSaveMeasurementFromOpenWeatherMap(
	ctx context.Context,
	appConfig *internal.AppConfig,
	runID uuid.UUID,
	station WeatherUnionDataStation,
) error {
	// initialize measurementID as UUID for open weather map
	measurementIDOpenWeatherMap, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	// get data from open weather map
	dataOpenWeatherMap, err := GetWeatherDataFromOpenWeatherMap(
		appConfig,
		station.Latitude,
		station.Longitude,
	)
	if err != nil {
		return err
	}

	// save data from open weather map
	err = SaveWeatherDataFromOpenWeatherMap(
		ctx,
		appConfig,
		measurementIDOpenWeatherMap,
		station.WeatherStationID,
		runID,
		&dataOpenWeatherMap,
	)
	if err != nil {
		fmt.Println("Error here 3!")
		return err
	}

	// return nil if all okay
	return nil
}

// function to save the measurement run ID
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
