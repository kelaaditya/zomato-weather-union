package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kelaaditya/zomato-weather-union/server/internal"
	"github.com/kelaaditya/zomato-weather-union/server/internal/utilities"
)

// structure of locality weather response from weather union
type WeatherUnionAPIReponseLocality struct {
	Status              *string                         `json:"status"`
	Message             *string                         `json:"message"`
	DeviceType          *int                            `json:"device_type"`
	LocalityWeatherData WeatherUnionDataWeatherLocality `json:"locality_weather_data"`
}

// structure of locality weather data from weather union
type WeatherUnionDataWeatherLocality struct {
	Temperature      *float64 `json:"temperature"`
	Humidity         *float64 `json:"humidity"`
	WindSpeed        *float64 `json:"wind_speed"`
	WindDirection    *float64 `json:"wind_direction"`
	RainIntensity    *float64 `json:"rain_intensity"`
	RainAccumulation *float64 `json:"rain_accumulation"`
}

// structure of weather station data from weather union
type WeatherUnionDataStation struct {
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
func GetWeatherDataFromWeatherUnionLocality(
	appConfig *internal.AppConfig,
	localityID string,
) (WeatherUnionAPIReponseLocality, error) {
	// base URL
	var baseURLString string = appConfig.ENVVariables.URLBaseWeatherUnion
	var localityPath string = "get_locality_weather_data"
	// join base URL with specific path
	specificURLString, err := url.JoinPath(baseURLString, localityPath)
	if err != nil {
		return WeatherUnionAPIReponseLocality{}, err
	}
	// parse to URL type
	specificURL, err := url.Parse(specificURLString)
	if err != nil {
		return WeatherUnionAPIReponseLocality{}, err
	}
	// add locality ID as the query parameter
	q := specificURL.Query()
	q.Set("locality_id", localityID)
	specificURL.RawQuery = q.Encode()
	// get string of the built API URL
	var APIURLString string = specificURL.String()

	// initialize new GET request
	request, err := http.NewRequest(http.MethodGet, APIURLString, nil)
	if err != nil {
		return WeatherUnionAPIReponseLocality{}, err
	}
	// add zomato API key to header
	request.Header.Add("X-Zomato-Api-Key", appConfig.ENVVariables.APIKeyWeatherUnion)
	// initialize new http client
	var client http.Client = http.Client{}
	// carry out get request
	response, err := client.Do(request)
	if err != nil {
		return WeatherUnionAPIReponseLocality{}, err
	}
	defer response.Body.Close()

	// check the http status codes
	if response.StatusCode != http.StatusOK {
		return WeatherUnionAPIReponseLocality{}, fmt.Errorf(
			"error in getting local data from weather union api. status: %v",
			response.StatusCode,
		)
	}

	// decode to JSON
	var localityAPIReponseObject WeatherUnionAPIReponseLocality
	// create JSON decoder
	jsonDecoder := json.NewDecoder(response.Body)
	if err := jsonDecoder.Decode(&localityAPIReponseObject); err != nil {
		return WeatherUnionAPIReponseLocality{}, err
	}

	return localityAPIReponseObject, nil
}

// function to store the Weather Union API call response to PostgreSQL
// measurementID:	each individual measurement
// runID:			each run (multiple individual measurements) will have one ID
func SaveWeatherDataFromWeatherUnion(
	ctx context.Context,
	appConfig *internal.AppConfig,
	measurementID uuid.UUID,
	weatherStationID uuid.UUID,
	runID uuid.UUID,
	data *WeatherUnionAPIReponseLocality,
) error {
	// postgresql query string
	var queryString string = `
	INSERT INTO measurements_weather_union(
		measurement_id,
		weather_station_id,
		run_id,
		message,
		device_type,
		temperature,
		humidity,
		wind_speed,
		wind_direction,
		rain_intensity,
		rain_accumulation
	)
	VALUES (
		@measurementID,
		@weatherStationID,
		@runID,
		@message,
		@deviceType,
		@temperature,
		@humidity,
		@windSpeed,
		@windDirection,
		@rainIntensity,
		@rainAccumulation
	);
	`

	// named arguments for building the query string
	var queryArguments pgx.NamedArgs = pgx.NamedArgs{
		"measurementID":    measurementID,
		"weatherStationID": weatherStationID,
		"runID":            runID,
		"message":          utilities.DereferenceOrNil(data.Message),
		"deviceType":       utilities.DereferenceOrNil(data.DeviceType),
		"temperature":      utilities.DereferenceOrNil(data.LocalityWeatherData.Temperature),
		"humidity":         utilities.DereferenceOrNil(data.LocalityWeatherData.Humidity),
		"windSpeed":        utilities.DereferenceOrNil(data.LocalityWeatherData.WindSpeed),
		"windDirection":    utilities.DereferenceOrNil(data.LocalityWeatherData.WindDirection),
		"rainIntensity":    utilities.DereferenceOrNil(data.LocalityWeatherData.RainIntensity),
		"rainAccumulation": utilities.DereferenceOrNil(data.LocalityWeatherData.RainAccumulation),
	}

	// executing the query string with the named arguments
	_, err := appConfig.DBPool.Exec(ctx, queryString, queryArguments)
	if err != nil {
		return fmt.Errorf(
			"error in inserting weather union data into postgresql: %w",
			err,
		)
	}

	return nil
}

// function to get Weather Union weather stations data (all)
func GetWeatherStationDataFromWeatherUnion(
	ctx context.Context,
	appConfig *internal.AppConfig,
) ([]WeatherUnionDataStation, error) {
	// placeholder slice
	var stationDataSlice []WeatherUnionDataStation

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
	LIMIT 5;
	`

	// prepare the query
	rows, err := appConfig.DBPool.Query(ctx, queryString)
	if err != nil {
		return nil, err
	}

	// run the query and collect rows
	stationDataSlice, err = pgx.CollectRows(
		rows,
		pgx.RowToStructByName[WeatherUnionDataStation],
	)
	if err != nil {
		return nil, err
	}

	// return slice of weather station structs
	// and nil error
	return stationDataSlice, nil
}
