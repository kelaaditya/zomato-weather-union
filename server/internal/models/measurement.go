package models

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kelaaditya/zomato-weather-union/server/internal"
	"golang.org/x/sync/errgroup"
)

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

// carry out a single run of measurements over all the
// weather stations from weather union
func GetAndSaveMeasurementsFromAPISingleRun(
	ctx context.Context,
	appConfig *internal.AppConfig,
) error {
	// initialize runID as UUID for this calculation
	runID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	// log runID when started
	appConfig.Logger.Info("run started.", "runID", runID.String())

	// get all weather station data from weather union
	sliceStationsWeatherUnion, err := GetWeatherStationsAllWeatherUnion(
		ctx,
		appConfig,
	)
	if err != nil {
		return err
	}

	// create a slice to append measurements from weather union
	var sliceMeasurementsWeatherUnion []WeatherUnionMeasurement
	// create a slice to append measurements from weather union
	var sliceMeasurementsOpenWeatherMap []OpenWeatherMapMeasurement

	// create a wait group
	var wgMeasurements errgroup.Group
	// create a mutex object
	var mutex sync.Mutex

	// iterate over all stations
	for _, station := range sliceStationsWeatherUnion {
		wgMeasurements.Go(func() error {
			// carry out API call to weather union
			measurementWeatherUnion, err := CallAPIWeatherUnionLocality(
				appConfig,
				&station,
				runID,
			)
			if err != nil {
				return err
			}

			// carry out API call to open weather map
			measurementOpenWeatherMap, err := CallAPIOpenWeatherMap(
				appConfig,
				&station,
				runID,
			)
			if err != nil {
				return err
			}

			// append new measurements to corresponding slices
			// lock and unlock slices while appending
			mutex.Lock()
			sliceMeasurementsWeatherUnion = append(
				sliceMeasurementsWeatherUnion,
				measurementWeatherUnion,
			)
			sliceMeasurementsOpenWeatherMap = append(
				sliceMeasurementsOpenWeatherMap,
				measurementOpenWeatherMap,
			)
			mutex.Unlock()

			// return nil if okay
			return nil
		})
	}

	// log the count of measurements received from weather union
	appConfig.Logger.Info(
		"measurements gathered from weather union",
		"total",
		strconv.Itoa(len(sliceMeasurementsWeatherUnion)),
	)
	// log the count of measurements received from open weather map
	appConfig.Logger.Info(
		"measurements gathered from open weather map",
		"total",
		strconv.Itoa(len(sliceMeasurementsOpenWeatherMap)),
	)

	// wait until all goroutines are completed
	// return the first non-nil error
	if err := wgMeasurements.Wait(); err != nil {
		// do not return an error here
		// as we want to continue the flag setting for those that
		// have been processed
		appConfig.Logger.Error(err.Error())
	}

	// save run ID
	err = SaveMeasurementRun(ctx, appConfig, runID)
	if err != nil {
		return err
	}

	// save measurements from weather union
	err = SaveMeasurementsWeatherUnion(
		ctx,
		appConfig,
		sliceMeasurementsWeatherUnion,
	)
	if err != nil {
		return err
	}

	// save measurement from open weather map
	err = SaveMeasurementOpenWeatherMap(
		ctx,
		appConfig,
		sliceMeasurementsOpenWeatherMap,
	)
	if err != nil {
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

// get all the unprocessed measurement values
// from both weather union and open weather map
func GetUnprocessedDataForCalculationsTemperature(
	ctx context.Context,
	appConfig *internal.AppConfig,
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

	// prepare the query
	rows, err := appConfig.DBPool.Query(ctx, queryString)
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
