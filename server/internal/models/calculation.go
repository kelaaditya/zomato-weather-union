package models

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kelaaditya/zomato-weather-union/server/internal"
	"golang.org/x/sync/errgroup"
)

// this type contains all parameters that are needed in
// one row entry in the calculations_temperature table
type CalculationTemperature struct {
	CalculationID               uuid.UUID
	MeasurementIDWeatherUnion   uuid.UUID
	MeasurementIDOpenWeatherMap uuid.UUID
	Method                      string
	DewPointTemperature         float64
	WetBulbTemperature          float64
}

// struct holding dew point and wet bulb temperature calculations
type Temperature struct {
	DewPointTemperature float64 `json:"temperature_dew_point"`
	WetBulbTemperature  float64 `json:"temperature_wet_bulb"`
}

// calculate all unprocessed wet bulb temperature values
// get the unprocessed data and then return a slice containing the values
func CalculateAndSaveTemperaturesAllUnprocessed(
	ctx context.Context,
	appConfig *internal.AppConfig,
) error {
	// get all unprocessed measurements
	sliceMeasurementsUnprocessed, err := GetUnprocessedDataForCalculationsTemperature(
		ctx,
		appConfig,
	)
	if err != nil {
		return err
	}

	// create a slice to append calculations to
	var sliceCalculationsSuccessful []CalculationTemperature

	// create a wait group
	var wgCalculations errgroup.Group
	// create a mutex object
	var mutex sync.Mutex

	// iterate over measurements
	for _, measurement := range sliceMeasurementsUnprocessed {
		wgCalculations.Go(func() error {
			// carry out calculations over a single measurement
			calculation, err := CalculateTemperatureSinglePass(measurement)
			if err != nil {
				return err
			}

			// append new successful calculation to slice of all successful
			// calculations
			// lock and unlock slice while appending
			mutex.Lock()
			sliceCalculationsSuccessful = append(
				sliceCalculationsSuccessful,
				calculation,
			)
			mutex.Unlock()

			// return nil if okay
			return nil
		})
	}

	// wait until all goroutines are completed
	// return the first non-nil error
	if err := wgCalculations.Wait(); err != nil {
		// do not return an error here
		// as we want to continue the flag setting for those that
		// have been processed
		appConfig.Logger.Error(err.Error())
	}

	// save calculations
	err = SaveCalculationsTemperatures(
		ctx,
		appConfig,
		sliceCalculationsSuccessful,
	)
	if err != nil {
		return err
	}

	// set flag is_processed for weather union measurements
	err = SetFlagsTemperature(
		ctx,
		appConfig,
		sliceMeasurementsUnprocessed,
		sliceCalculationsSuccessful,
	)
	if err != nil {
		return err
	}

	// return nil if all okay
	return nil
}

// run the python script from the fetched
// data of one instance
func CalculateTemperatureSinglePass(
	measurement MeasurementTemperature,
) (CalculationTemperature, error) {
	// initialize calculationID as UUID for this calculation
	calculationID, err := uuid.NewRandom()
	if err != nil {
		return CalculationTemperature{}, err
	}

	// placeholder for temperature struct
	var temperature Temperature

	// build command line argument strings
	var CLATemperature = "--temperature=" + fmt.Sprint(measurement.Temperature)
	var CLAHumidity = "--humidity=" + fmt.Sprint(measurement.Humidity)
	var CLAPressure = "--pressure=" + fmt.Sprint(measurement.Pressure)

	// create os command
	// to debug the command, do this
	// fmt.Println(command.String())
	command := exec.Command(
		"python3",
		"../scripts/wet_bulb_temperature.py",
		CLATemperature,
		CLAHumidity,
		CLAPressure,
	)

	// get the stdout after running the command
	stdout, err := command.Output()
	if err != nil {
		return CalculationTemperature{}, err
	}

	// convert the JSON slice of bytes to struct
	err = json.Unmarshal(stdout, &temperature)
	if err != nil {
		return CalculationTemperature{}, err
	}

	// create calculation struct for entry in database
	calculation := CalculationTemperature{
		CalculationID:               calculationID,
		MeasurementIDWeatherUnion:   measurement.MeasurementIDWeatherUnion,
		MeasurementIDOpenWeatherMap: measurement.MeasurementIDOpenWeatherMap,
		Method:                      "metpy-with-open-weather-map",
		DewPointTemperature:         temperature.DewPointTemperature,
		WetBulbTemperature:          temperature.WetBulbTemperature,
	}

	return calculation, nil
}

// save the successful wet bulb and dew point temperature calculations
// to the database
func SaveCalculationsTemperatures(
	ctx context.Context,
	appConfig *internal.AppConfig,
	sliceCalculationsSuccessful []CalculationTemperature,
) error {
	// create a slice containing a slice of any types
	// length of the holder slice is the length of calculations
	// passed to the function
	var sliceInsertValues [][]any = make(
		[][]any,
		len(sliceCalculationsSuccessful),
	)

	// build slice of values for bulk insert
	for i, calculation := range sliceCalculationsSuccessful {
		sliceInsertValues[i] = []any{
			calculation.CalculationID,
			calculation.MeasurementIDWeatherUnion,
			calculation.Method,
			calculation.DewPointTemperature,
			calculation.WetBulbTemperature,
		}
	}

	// create a bulk insert query
	_, err := appConfig.DBPool.CopyFrom(
		ctx,
		pgx.Identifier{"calculations_temperature"},
		[]string{
			"calculation_id",
			"measurement_id_weather_union",
			"method",
			"temperature_dew_point",
			"temperature_wet_bulb",
		},
		pgx.CopyFromRows(sliceInsertValues),
	)
	if err != nil {
		return fmt.Errorf(
			"error in inserting calculation into postgresql: %w",
			err,
		)
	}

	// return nil if all okay
	return nil
}

// function to set is_processed and is_successful
// for weather union and open weather map
func SetFlagsTemperature(
	ctx context.Context,
	appConfig *internal.AppConfig,
	sliceMeasurementsUnprocessed []MeasurementTemperature,
	sliceCalculationsSuccessful []CalculationTemperature,
) error {
	// create batch updates for postgresql entry
	var queryBatch *pgx.Batch = &pgx.Batch{}

	// unprocessed measurements
	// slice to gather all measurement IDs from weather union
	var sliceMeasurementIDsUnprocessedWeatherUnion []uuid.UUID
	// slice to gather all measurement IDs from open weather map
	var sliceMeasurementIDsUnprocessedOpenWeatherMap []uuid.UUID
	// iterate over unprocessed measurements
	for _, measurement := range sliceMeasurementsUnprocessed {
		// append measurement IDs of weather union
		// from unprocessed measurements
		sliceMeasurementIDsUnprocessedWeatherUnion = append(
			sliceMeasurementIDsUnprocessedWeatherUnion,
			measurement.MeasurementIDWeatherUnion,
		)
		// append measurement IDs of open weather map
		// from unprocessed measurements
		sliceMeasurementIDsUnprocessedOpenWeatherMap = append(
			sliceMeasurementIDsUnprocessedOpenWeatherMap,
			measurement.MeasurementIDOpenWeatherMap,
		)
	}

	// successful calculations
	// slice to gather all measurement IDs from weather union
	var sliceMeasurementIDsSuccessfulWeatherUnion []uuid.UUID
	// slice to gather all measurement IDs from open weather map
	var sliceMeasurementIDsSuccessfulOpenWeatherMap []uuid.UUID
	// iterate over successful calculations
	for _, calculation := range sliceCalculationsSuccessful {
		// append measurement IDs of weather union
		// from successful calculations
		sliceMeasurementIDsSuccessfulWeatherUnion = append(
			sliceMeasurementIDsSuccessfulWeatherUnion,
			calculation.MeasurementIDWeatherUnion,
		)
		// append measurement IDs of open weather map
		// from successfull calculations
		sliceMeasurementIDsSuccessfulOpenWeatherMap = append(
			sliceMeasurementIDsSuccessfulOpenWeatherMap,
			calculation.MeasurementIDOpenWeatherMap,
		)
	}

	// flag is_processed
	// weather union
	// postgresql query string
	var queryStringIsProcessedWeatherUnion string = `
	UPDATE measurements_weather_union
	SET is_processed_for_calculation_temperature = TRUE
	WHERE measurement_id = ANY(@arrayMeasurementIDs);
	`
	// named arguments for building the query string
	var queryArgumentsIsProcessedWeatherUnion pgx.NamedArgs = pgx.NamedArgs{
		"arrayMeasurementIDs": sliceMeasurementIDsUnprocessedWeatherUnion,
	}
	// append to pg query batch
	queryBatch.Queue(
		queryStringIsProcessedWeatherUnion,
		queryArgumentsIsProcessedWeatherUnion,
	)

	// flag is_processed
	// open weather map
	// postgresql query string
	var queryStringIsProcessedOpenWeatherMap string = `
	UPDATE measurements_open_weather_map
	SET is_processed_for_calculation_temperature = TRUE
	WHERE measurement_id = ANY(@arrayMeasurementIDs);
	`
	// named arguments for building the query string
	var queryArgumentsIsProcessedOpenWeatherMap pgx.NamedArgs = pgx.NamedArgs{
		"arrayMeasurementIDs": sliceMeasurementIDsUnprocessedOpenWeatherMap,
	}
	// append to pg query batch
	queryBatch.Queue(
		queryStringIsProcessedOpenWeatherMap,
		queryArgumentsIsProcessedOpenWeatherMap,
	)

	// flag is_successful
	// weather union
	// postgresql query string
	var queryStringIsSuccessfulWeatherUnion string = `
	UPDATE measurements_weather_union
	SET is_successful_for_calculation_temperature = TRUE
	WHERE measurement_id = ANY(@arrayMeasurementIDs);
	`
	// named arguments for building the query string
	var queryArgumentsIsSuccessfulWeatherUnion pgx.NamedArgs = pgx.NamedArgs{
		"arrayMeasurementIDs": sliceMeasurementIDsSuccessfulWeatherUnion,
	}
	// append to pg query batch
	queryBatch.Queue(
		queryStringIsSuccessfulWeatherUnion,
		queryArgumentsIsSuccessfulWeatherUnion,
	)

	// flag is_successful
	// open weather map
	// postgresql query string
	var queryStringIsSuccessfulOpenWeatherMap string = `
	UPDATE measurements_open_weather_map
	SET is_successful_for_calculation_temperature = TRUE
	WHERE measurement_id = ANY(@arrayMeasurementIDs);
	`
	// named arguments for building the query string
	var queryArgumentsIsSuccessfulOpenWeatherMap pgx.NamedArgs = pgx.NamedArgs{
		"arrayMeasurementIDs": sliceMeasurementIDsSuccessfulOpenWeatherMap,
	}
	// append to pg query batch
	queryBatch.Queue(
		queryStringIsSuccessfulOpenWeatherMap,
		queryArgumentsIsSuccessfulOpenWeatherMap,
	)

	// send the batch query via the connection pool
	var batchResults pgx.BatchResults = appConfig.DBPool.SendBatch(
		ctx,
		queryBatch,
	)

	// execute all batches
	// (is_proc_wu, is_proc_owm, is_succ_wu, is_succ_owm)
	// error placeholder
	_, err := batchResults.Exec()
	if err != nil {
		return fmt.Errorf(
			"error in setting flag into postgresql: %w",
			err,
		)
	}
	_, err = batchResults.Exec()
	if err != nil {
		return fmt.Errorf(
			"error in setting flag into postgresql: %w",
			err,
		)
	}
	_, err = batchResults.Exec()
	if err != nil {
		return fmt.Errorf(
			"error in setting flag into postgresql: %w",
			err,
		)
	}
	_, err = batchResults.Exec()
	if err != nil {
		return fmt.Errorf(
			"error in setting flag into postgresql: %w",
			err,
		)
	}

	// close the batch connection
	err = batchResults.Close()
	if err != nil {
		return fmt.Errorf(
			"error in setting flag into postgresql: %w",
			err,
		)
	}

	// return nil if all okay
	return nil
}
