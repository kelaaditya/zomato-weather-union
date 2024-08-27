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
)

// structure of locality weather reponse from weather union
type WeatherUnionLocalityAPIReponse struct {
	Status              int                             `json:"status"`
	Message             string                          `json:"message"`
	DeviceType          int                             `json:"device_type"`
	LocalityWeatherData WeatherUnionLocalityWeatherData `json:"locality_weather_data"`
}

// structure of locality weather data from weather union
type WeatherUnionLocalityWeatherData struct {
	Temperature      float64 `json:"temperature"`
	Humidity         float64 `json:"humidity"`
	WindSpeed        float64 `json:"wind_speed"`
	WindDirection    float64 `json:"wind_direction"`
	RainIntensity    float64 `json:"rain_intensity"`
	RainAccumulation float64 `json:"rain_accumulation"`
}

// get weather data from locality (weather union)
func GetWeatherDataFromWeatherUnionLocality(
	appConfig *internal.AppConfig,
	localityID string,
	c chan<- WeatherUnionLocalityAPIReponse,
) (WeatherUnionLocalityAPIReponse, error) {
	// build the API URL
	var baseAPI string = appConfig.ENVVariables.URLAPIWeatherUnion
	var localityPath string = "get_locality_weather_data"
	localityAPI, err := url.Parse(baseAPI + localityPath)
	if err != nil {
		return WeatherUnionLocalityAPIReponse{}, err
	}
	// add locality ID as the query parameter
	q := localityAPI.Query()
	q.Set("locality_id", localityID)
	localityAPI.RawQuery = q.Encode()

	// carry out get request
	response, err := http.Get(localityAPI.RawPath)
	if err != nil {
		return WeatherUnionLocalityAPIReponse{}, err
	}
	defer response.Body.Close()

	// check the http status codes
	if response.StatusCode != http.StatusOK {
		return WeatherUnionLocalityAPIReponse{}, fmt.Errorf(
			"error in getting local data from weather union api. status: %v",
			response.StatusCode,
		)
	}

	var localityAPIReponseObject WeatherUnionLocalityAPIReponse
	// decode to JSON
	jsonDecoder := json.NewDecoder(response.Body)
	if err := jsonDecoder.Decode(&localityAPIReponseObject); err != nil {
		return WeatherUnionLocalityAPIReponse{}, err
	}

	return localityAPIReponseObject, nil
}

// function to store the Weather Union API call response to PostgreSQL
func StoreWeatherDataFromWeatherUnion(
	ctx context.Context,
	appConfig *internal.AppConfig,
	measurementID uuid.UUID,
	weatherStationID string,
	data *WeatherUnionLocalityAPIReponse,
) error {
	// postgresql query string
	var queryString string = `
	INSERT INTO measurements_weather_union(
		measurement_id,
		weather_station_id,
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
		"temperature":      data.LocalityWeatherData.Temperature,
		"humidity":         data.LocalityWeatherData.Humidity,
		"windSpeed":        data.LocalityWeatherData.WindSpeed,
		"windDirection":    data.LocalityWeatherData.WindDirection,
		"rainIntensity":    data.LocalityWeatherData.RainIntensity,
		"rainAccumulation": data.LocalityWeatherData.RainAccumulation,
	}

	// executing the query string with the named arguments
	_, err := appConfig.DBPool.Exec(ctx, queryString, queryArguments)
	if err != nil {
		return fmt.Errorf("error in inserting weather union data into postgresql: %w", err)
	}

	return nil
}
