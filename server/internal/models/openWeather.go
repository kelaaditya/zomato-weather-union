package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kelaaditya/zomato-weather-union/server/internal"
	"github.com/kelaaditya/zomato-weather-union/server/internal/utilities"
)

// type to hold a "measurement" from open weather map
// see the schema structure for the table "measurements_open_weather_map"
// in the PostgreSQL migration files.
type OpenWeatherMapMeasurement struct {
	MeasurementID            uuid.UUID `json:"measurement_id"`
	WeatherStationID         uuid.UUID `json:"weather_station_id"`
	RunID                    uuid.UUID `json:"run_id"`
	OpenWeatherMapAPIReponse           // embedded struct
}

// structure of weather reponse from open weather map
type OpenWeatherMapAPIReponse struct {
	Latitude       *float64                           `json:"lat"`
	Longitude      *float64                           `json:"lon"`
	TimeZone       *string                            `json:"timezone"`
	TimeZoneOffset *int64                             `json:"timezone_offset"`
	Current        OpenWeatherMapObjectWeatherCurrent `json:"current"`
}

// structure of current weather data from open weather map
type OpenWeatherMapObjectWeatherCurrent struct {
	TimeCurrent   *int64                           `json:"dt"`
	TimeSunrise   *int64                           `json:"sunrise"`
	TimeSunset    *int64                           `json:"sunset"`
	Temperature   *float64                         `json:"temp"`
	FeelsLike     *float64                         `json:"feels_like"`
	Pressure      *float64                         `json:"pressure"`
	Humidity      *float64                         `json:"humidity"`
	DewPoint      *float64                         `json:"dew_point"`
	UVIndex       *float64                         `json:"uvi"`
	Clouds        *float64                         `json:"clouds"`
	Visibility    *int64                           `json:"visibility"`
	WindSpeed     *float64                         `json:"wind_speed"`
	WindDirection *float64                         `json:"wind_deg"`
	WindGust      *float64                         `json:"wind_gust"`
	WeatherObject []OpenWeatherMapObjectSubWeather `json:"weather"`
}

// structure of the weather object inside the current weather
// data from open weather map
type OpenWeatherMapObjectSubWeather struct {
	ID          *int    `json:"id"`
	Main        *string `json:"main"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}

// function for calling the Open Weather API
func CallAPIOpenWeatherMap(
	appConfig *internal.AppConfig,
	station *WeatherUnionStation,
	runID uuid.UUID,
) (OpenWeatherMapMeasurement, error) {
	// initialize measurementID as UUID for this calculation
	measurementID, err := uuid.NewRandom()
	if err != nil {
		return OpenWeatherMapMeasurement{}, err
	}

	//
	// build URL string for API call
	//
	// query parameters map
	var mapQueryParameters map[string]string = map[string]string{
		"lat":     station.Latitude,
		"lon":     station.Longitude,
		"exclude": "minutely,hourly,daily,alerts",
		"appid":   appConfig.ENVVariables.APIKeyOpenWeatherMap,
	}
	// call utility function to build URL string
	URLString, err := utilities.BuildURLString(
		appConfig.ENVVariables.URLBaseOpenWeatherMap,
		"onecall",
		mapQueryParameters,
	)
	if err != nil {
		return OpenWeatherMapMeasurement{}, nil
	}

	//
	// carry out GET request
	//
	response, err := http.Get(URLString)
	if err != nil {
		return OpenWeatherMapMeasurement{}, err
	}
	defer response.Body.Close()

	// check the http status codes
	if response.StatusCode != http.StatusOK {
		return OpenWeatherMapMeasurement{}, fmt.Errorf(
			"error in getting weather from open weather maps api. status: %v",
			response.StatusCode,
		)
	}

	//
	// decode JSON
	//
	// holder for data from decode
	var localityAPIReponseObject OpenWeatherMapAPIReponse
	// create JSON decoder
	jsonDecoder := json.NewDecoder(response.Body)
	if err := jsonDecoder.Decode(&localityAPIReponseObject); err != nil {
		return OpenWeatherMapMeasurement{}, err
	}

	//
	// initialize the measurement object for open weather map
	//
	var measurement OpenWeatherMapMeasurement = OpenWeatherMapMeasurement{
		MeasurementID:            measurementID,
		WeatherStationID:         station.WeatherStationID,
		RunID:                    runID,
		OpenWeatherMapAPIReponse: localityAPIReponseObject,
	}

	return measurement, nil
}

// function to store a slice of measurements to the database
// in bulk insert
func SaveMeasurementOpenWeatherMap(
	ctx context.Context,
	appConfig *internal.AppConfig,
	sliceMeasurementsOpenWeatherMap []OpenWeatherMapMeasurement,
) error {
	// create a slice containing a slice of any types
	// length of the holder slice is the length of calculations
	// passed to the function
	var sliceInsertValues [][]any = make(
		[][]any,
		len(sliceMeasurementsOpenWeatherMap),
	)

	// build a slice of values for bulk insert
	for i, measurement := range sliceMeasurementsOpenWeatherMap {
		sliceInsertValues[i] = []any{
			measurement.MeasurementID,
			measurement.WeatherStationID,
			measurement.RunID,
			utilities.DereferenceOrNil(measurement.TimeZone),
			utilities.DereferenceOrNil(measurement.TimeZoneOffset),
			utilities.DereferenceOrNil(measurement.Current.TimeCurrent),
			utilities.DereferenceOrNil(measurement.Current.TimeSunrise),
			utilities.DereferenceOrNil(measurement.Current.TimeSunset),
			utilities.DereferenceOrNil(measurement.Current.Temperature),
			utilities.DereferenceOrNil(measurement.Current.FeelsLike),
			utilities.DereferenceOrNil(measurement.Current.Pressure),
			utilities.DereferenceOrNil(measurement.Current.Humidity),
			utilities.DereferenceOrNil(measurement.Current.DewPoint),
			utilities.DereferenceOrNil(measurement.Current.UVIndex),
			utilities.DereferenceOrNil(measurement.Current.Clouds),
			utilities.DereferenceOrNil(measurement.Current.Visibility),
			utilities.DereferenceOrNil(measurement.Current.WindSpeed),
			utilities.DereferenceOrNil(measurement.Current.WindDirection),
			utilities.DereferenceOrNil(measurement.Current.WindGust),
			utilities.DereferenceOrNil(measurement.Current.WeatherObject[0].ID),
			utilities.DereferenceOrNil(measurement.Current.WeatherObject[0].Main),
			utilities.DereferenceOrNil(measurement.Current.WeatherObject[0].Description),
			utilities.DereferenceOrNil(measurement.Current.WeatherObject[0].Icon),
		}
	}

	// create a bulk insert query
	_, err := appConfig.DBPool.CopyFrom(
		ctx,
		pgx.Identifier{"measurements_open_weather_map"},
		[]string{
			"measurement_id",
			"weather_station_id",
			"run_id",
			"time_zone",
			"time_zone_offset",
			"time_current",
			"time_sunrise",
			"time_sunset",
			"temperature",
			"feels_like",
			"pressure",
			"humidity",
			"dew_point",
			"uv_index",
			"clouds",
			"visibility",
			"wind_speed",
			"wind_direction",
			"wind_gust",
			"weather_object_id",
			"weather_object_main",
			"weather_object_description",
			"weather_object_icon",
		},
		pgx.CopyFromRows(sliceInsertValues),
	)
	if err != nil {
		return fmt.Errorf(
			"error in inserting measurement into postgresql: %w",
			err,
		)
	}

	return nil
}
