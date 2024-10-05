package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelaaditya/zomato-weather-union/server/internal/utilities"
)

// model struct for weather union
type WeatherUnionModel struct {
	DB *pgxpool.Pool
}

// type to hold a "measurement" from weather union
// see the schema structure for the table "measurement_weather_union"
// in the PostgreSQL migration files.
type WeatherUnionMeasurement struct {
	MeasurementID                  uuid.UUID `json:"measurement_id"`
	WeatherStationID               uuid.UUID `json:"weather_station_id"`
	RunID                          uuid.UUID `json:"run_id"`
	WeatherUnionAPIReponseLocality           // embedded struct
}

// structure of locality weather response from weather union
type WeatherUnionAPIReponseLocality struct {
	Status              *string                           `json:"status"`
	Message             *string                           `json:"message"`
	DeviceType          *int                              `json:"device_type"`
	LocalityWeatherData WeatherUnionObjectWeatherLocality `json:"locality_weather_data"`
}

// structure of locality weather data from weather union
type WeatherUnionObjectWeatherLocality struct {
	Temperature      *float64 `json:"temperature"`
	Humidity         *float64 `json:"humidity"`
	WindSpeed        *float64 `json:"wind_speed"`
	WindDirection    *float64 `json:"wind_direction"`
	RainIntensity    *float64 `json:"rain_intensity"`
	RainAccumulation *float64 `json:"rain_accumulation"`
}

// structure of weather station data from weather union
type WeatherUnionStation struct {
	WeatherStationID  uuid.UUID `json:"weather_station_id"`
	CityName          string    `json:"city_name"`
	LocalityName      string    `json:"locality_name"`
	LocalityID        string    `json:"locality_id"`
	Latitude          string    `json:"latitude"`
	Longitude         string    `json:"longitude"`
	DeviceType        string    `json:"device_type"`
	DeviceTypeInteger int       `json:"device_type_integer"`
}

// get weather data from locality (weather union)
func (model WeatherUnionModel) CallAPIWeatherUnionLocality(
	APIBaseURL string,
	APIKey string,
	station *WeatherUnionStation,
	runID uuid.UUID,
) (WeatherUnionMeasurement, error) {
	// initialize measurementID as UUID for this calculation
	measurementID, err := uuid.NewRandom()
	if err != nil {
		return WeatherUnionMeasurement{}, err
	}

	//
	// build URL string for API call
	//
	// query parameters map
	var mapQueryParameters map[string]string = map[string]string{
		"locality_id": station.LocalityID,
	}
	// call utility function to build URL string
	URLString, err := utilities.BuildURLString(
		APIBaseURL,
		"get_locality_weather_data",
		mapQueryParameters,
	)
	if err != nil {
		return WeatherUnionMeasurement{}, err
	}

	//
	// initialize new GET request
	//
	// new request
	request, err := http.NewRequest(http.MethodGet, URLString, nil)
	if err != nil {
		return WeatherUnionMeasurement{}, err
	}
	// add zomato API key to header
	request.Header.Add(
		"X-Zomato-Api-Key",
		APIKey,
	)
	// initialize new http client
	var client http.Client = http.Client{}
	// carry out get request
	response, err := client.Do(request)
	if err != nil {
		return WeatherUnionMeasurement{}, err
	}
	defer response.Body.Close()

	// check the http status codes
	if response.StatusCode != http.StatusOK {
		return WeatherUnionMeasurement{}, fmt.Errorf(
			"error in getting local data from weather union api. status: %v",
			response.StatusCode,
		)
	}

	//
	// decode JSON
	//
	// holder for data from decode
	var localityAPIReponseObject WeatherUnionAPIReponseLocality
	// create JSON decoder
	jsonDecoder := json.NewDecoder(response.Body)
	if err := jsonDecoder.Decode(&localityAPIReponseObject); err != nil {
		return WeatherUnionMeasurement{}, err
	}

	//
	// initialize the measurement object for weather union
	//
	var measurement WeatherUnionMeasurement = WeatherUnionMeasurement{
		MeasurementID:                  measurementID,
		WeatherStationID:               station.WeatherStationID,
		RunID:                          runID,
		WeatherUnionAPIReponseLocality: localityAPIReponseObject,
	}

	return measurement, nil
}

// function to store a slice of measurements to the database
// in one bulk insert
func (model WeatherUnionModel) SaveMeasurementsWeatherUnion(
	ctx context.Context,
	sliceMeasurementsWeatherUnion []WeatherUnionMeasurement,
) error {
	// create a slice containing a slice of any types
	// length of the holder slice is the length of calculations
	// passed to the function
	var sliceInsertValues [][]any = make(
		[][]any,
		len(sliceMeasurementsWeatherUnion),
	)

	// build a slice of values for bulk insert
	for i, measurement := range sliceMeasurementsWeatherUnion {
		sliceInsertValues[i] = []any{
			measurement.MeasurementID,
			measurement.WeatherStationID,
			measurement.RunID,
			utilities.DereferenceOrNil(measurement.Message),
			utilities.DereferenceOrNil(measurement.DeviceType),
			utilities.DereferenceOrNil(measurement.LocalityWeatherData.Temperature),
			utilities.DereferenceOrNil(measurement.LocalityWeatherData.Humidity),
			utilities.DereferenceOrNil(measurement.LocalityWeatherData.WindSpeed),
			utilities.DereferenceOrNil(measurement.LocalityWeatherData.WindDirection),
			utilities.DereferenceOrNil(measurement.LocalityWeatherData.RainIntensity),
			utilities.DereferenceOrNil(measurement.LocalityWeatherData.RainAccumulation),
		}
	}

	// create a 5 second timeout context
	ctxWT, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancellation of the timeout
	defer cancel()

	// create a bulk insert query
	_, err := model.DB.CopyFrom(
		ctxWT,
		pgx.Identifier{"measurements_weather_union"},
		[]string{
			"measurement_id",
			"weather_station_id",
			"run_id",
			"message",
			"device_type",
			"temperature",
			"humidity",
			"wind_speed",
			"wind_direction",
			"rain_intensity",
			"rain_accumulation",
		},
		pgx.CopyFromRows(sliceInsertValues),
	)
	if err != nil {
		return fmt.Errorf(
			"error in inserting measurement into postgresql: %w",
			err,
		)
	}

	// return nil if all okay
	return nil
}

// function to get Weather Union weather stations data (all)
func (model WeatherUnionModel) GetWeatherStationsAllWeatherUnion(
	ctx context.Context,
) (
	[]WeatherUnionStation,
	error,
) {
	// placeholder slice
	var stationDataSlice []WeatherUnionStation

	// query string
	var queryString string = `
	SELECT
		weather_station_id,
		city_name,
		locality_name,
		locality_id,
		ST_X(location::geometry) AS longitude,
		ST_Y(location::geometry) AS latitude,
		device_type,
		device_type_integer
	FROM weather_union_stations
	LIMIT 10;
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
	stationDataSlice, err = pgx.CollectRows(
		rows,
		pgx.RowToStructByName[WeatherUnionStation],
	)
	if err != nil {
		return nil, err
	}

	// return slice of weather station structs
	// and nil error
	return stationDataSlice, nil
}
