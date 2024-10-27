package models

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// model struct for calculations
type CalculationModel struct {
	DB *pgxpool.Pool
}

// this type contains all parameters that are needed in
// one row entry in the calculations_temperature table
type CalculationTemperature struct {
	CalculationID               uuid.UUID
	MeasurementIDWeatherUnion   uuid.UUID
	MeasurementIDOpenWeatherMap uuid.UUID
	Method                      string
	TemperatureDewPoint         float64
	TemperatureWetBulb          float64
}

// struct holding dew point and wet bulb temperature calculations
type Temperature struct {
	DewPoint float64 `json:"temperature_dew_point"`
	WetBulb  float64 `json:"temperature_wet_bulb"`
}

// type to hold relevant data for display on front end
type CalculationTemperatureWithStationDetails struct {
	RunID                uuid.UUID `db:"run_id" json:"run_id"`
	LocalityID           string    `db:"locality_id" json:"locality_id"`
	LocalityName         string    `db:"locality_name" json:"locality_name"`
	Latitude             string    `db:"latitude" json:"latitude"`
	Longitude            string    `db:"longitude" json:"longitude"`
	TemperatureDewPoint  float64   `db:"temperature_dew_point" json:"temperature_dew_point"`
	TemperatureWetBulb   float64   `db:"temperature_wet_bulb" json:"temperature_wet_bulb"`
	CalculationTimeStamp time.Time `db:"time_stamp_calculation" json:"time_stamp_calculation"`
}

// run the python script from the fetched
// data of one instance
func (model CalculationModel) CalculateTemperatureFromSingleMeasurement(
	pathToPythonEnvironment string,
	measurement MeasurementTemperature,
) (CalculationTemperature, error) {
	// initialize calculationID as UUID for this calculation
	calculationID, err := uuid.NewRandom()
	if err != nil {
		return CalculationTemperature{}, err
	}

	// join path of python environment along with the python3
	// executable
	environmentBinary := path.Join(pathToPythonEnvironment, "python3")

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
		environmentBinary,
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
		TemperatureDewPoint:         temperature.DewPoint,
		TemperatureWetBulb:          temperature.WetBulb,
	}

	return calculation, nil
}

// save the successful wet bulb and dew point temperature calculations
// to the database
func (model CalculationModel) SaveCalculationsTemperatures(
	ctx context.Context,
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
			calculation.TemperatureDewPoint,
			calculation.TemperatureWetBulb,
		}
	}

	// create a 5 second timeout context
	ctxWT, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancellation of the timeout
	defer cancel()

	// create a bulk insert query
	_, err := model.DB.CopyFrom(
		ctxWT,
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
func (model CalculationModel) SetFlagsTemperature(
	ctx context.Context,
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

	// create a 5 second timeout context
	ctxWT, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancellation of the timeout
	defer cancel()

	// send the batch query via the connection pool
	var batchResults pgx.BatchResults = model.DB.SendBatch(
		ctxWT,
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

// get the temperature calculations for display
// from a single run
func (model CalculationModel) GetCalculationsTemperatureWithStationDetails(
	ctx context.Context,
) (
	[]CalculationTemperatureWithStationDetails,
	error,
) {
	// placeholder slice
	var sliceCalculations []CalculationTemperatureWithStationDetails

	// query string
	var queryString string = `
	SELECT
		ROUND(ct.temperature_wet_bulb::NUMERIC, 3)::FLOAT
			AS temperature_wet_bulb,
		ROUND(ct.temperature_dew_point::NUMERIC, 3)::FLOAT
			AS temperature_dew_point,
		ct.time_stamp AS time_stamp_calculation,
		mwu.run_id,
		wus.locality_id,
		wus.locality_name,
		ST_X(wus.location::geometry) AS longitude,
		ST_Y(wus.location::geometry) AS latitude
	FROM calculations_temperature ct
	JOIN measurements_weather_union mwu
	ON ct.measurement_id_weather_union = mwu.measurement_id
	JOIN weather_union_stations wus
	ON mwu.weather_station_id = wus.weather_station_id
	WHERE run_id = (
		SELECT run_id
		FROM measurement_runs
		ORDER BY time_stamp DESC
		LIMIT 1
	)
	ORDER BY temperature_wet_bulb DESC;
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
	sliceCalculations, err = pgx.CollectRows(
		rows,
		pgx.RowToStructByName[CalculationTemperatureWithStationDetails],
	)
	if err != nil {
		return nil, err
	}

	// return slice of calculations for display
	return sliceCalculations, nil
}
