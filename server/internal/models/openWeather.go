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

// structure of weather reponse from open weather map
type OpenWeatherAPIReponse struct {
	Latitude       float64                       `json:"lat"`
	Longitude      float64                       `json:"lon"`
	TimeZone       string                        `json:"timezone"`
	TimeZoneOffset int64                         `json:"timezone_offset"`
	Current        OpenWeatherCurrentWeatherData `json:"current"`
}

type OpenWeatherCurrentWeatherData struct {
	TimeCurrent   int64                         `json:"dt"`
	TimeSunrise   int64                         `json:"sunrise"`
	TimeSunset    int64                         `json:"sunset"`
	Temperature   float64                       `json:"temp"`
	FeelsLike     float64                       `json:"feels_like"`
	Pressure      float64                       `json:"pressure"`
	Humidity      float64                       `json:"humidity"`
	DewPoint      float64                       `json:"dew_point"`
	UVIndex       float64                       `json:"uvi"`
	Clouds        float64                       `json:"clouds"`
	Visibility    int64                         `json:"visibility"`
	WindSpeed     float64                       `json:"wind_speed"`
	WindDirection float64                       `json:"wind_deg"`
	WindGust      float64                       `json:"wind_gust"`
	WeatherObject []OpenWeatherSubWeatherObject `json:"weather"`
}

type OpenWeatherSubWeatherObject struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// function for calling the Open Weather API
func GetWeatherDataFromOpenWeatherMap(
	appConfig *internal.AppConfig,
	latitude string,
	longitude string,
) (OpenWeatherAPIReponse, error) {
	// build the API URL
	var baseAPI string = appConfig.ENVVariables.URLAPIWeatherUnion
	var oneCallPath string = "onecall"
	weatherAPI, err := url.Parse(baseAPI + oneCallPath)
	if err != nil {
		return OpenWeatherAPIReponse{}, err
	}
	// add locality ID as the query parameter
	q := weatherAPI.Query()
	q.Add("lat", latitude)
	q.Add("lon", longitude)
	q.Add("exclude", "minutely,hourly,daily,alerts")
	q.Add("appid", appConfig.ENVVariables.APIKeyOpenWeatherMap)
	weatherAPI.RawQuery = q.Encode()

	// carry out get request
	response, err := http.Get(weatherAPI.RawPath)
	if err != nil {
		return OpenWeatherAPIReponse{}, err
	}
	defer response.Body.Close()

	// check the http status codes
	if response.StatusCode != http.StatusOK {
		return OpenWeatherAPIReponse{}, fmt.Errorf(
			"error in getting weather from open weather maps api. status: %v",
			response.StatusCode,
		)
	}

	var localityAPIReponseObject OpenWeatherAPIReponse
	// decode to JSON
	jsonDecoder := json.NewDecoder(response.Body)
	if err := jsonDecoder.Decode(&localityAPIReponseObject); err != nil {
		return OpenWeatherAPIReponse{}, err
	}

	return localityAPIReponseObject, nil
}

// function to store the Open Weather API call response to PostgreSQL
func StoreWeatherDataFromOpenWeatherMap(
	ctx context.Context,
	appConfig *internal.AppConfig,
	measurementID uuid.UUID,
	data *OpenWeatherAPIReponse,
) error {
	// postgresql query string
	var queryString string = `
	INSERT INTO measurements_open_weather_map(
		measurement_id,
		time_zone,
		time_zone_offset,
		time_current,
		time_sunrise,
		time_sunset,
		temperature,
		feels_like,
		pressure,
		humidity,
		dew_point,
		uv_index,
		clouds,
		visibility,
		wind_speed,
		wind_direction,
		wind_gust,
		weather_object_id,
		weather_object_main,
		weather_object_description,
		weather_object_icon
	)
	VALUES (
		@measurementID,
		@timeZone,
		@timeZoneOffset,
		@timeCurrent,
		@timeSunrise,
		@timeSunset,
		@temperature,
		@feelsLike,
		@pressure,
		@humidity,
		@dewPoint,
		@UVIndex,
		@clouds,
		@visibility,
		@windSpeed,
		@windDirection,
		@windGust,
		@weatherObjectID,
		@weatherObjectMain,
		@weatherObjectDescription,
		@weatherObjectIcon
	);
	`

	// named arguments for building the query string
	var queryArguments pgx.NamedArgs = pgx.NamedArgs{
		"measurementID":            measurementID,
		"timezone":                 data.TimeZone,
		"timezoneOffset":           data.TimeZoneOffset,
		"timeCurrent":              data.Current.TimeCurrent,
		"timeSunrise":              data.Current.TimeSunrise,
		"timeSunset":               data.Current.TimeSunset,
		"temperature":              data.Current.Temperature,
		"feelsLike":                data.Current.FeelsLike,
		"pressure":                 data.Current.Pressure,
		"humidity":                 data.Current.Humidity,
		"dewPoint":                 data.Current.DewPoint,
		"UVIndex":                  data.Current.UVIndex,
		"clouds":                   data.Current.Clouds,
		"visibility":               data.Current.Visibility,
		"windSpeed":                data.Current.WindSpeed,
		"windDirection":            data.Current.WindDirection,
		"windGust":                 data.Current.WindGust,
		"weatherObjectID":          data.Current.WeatherObject[0].ID,
		"weatherObjectMain":        data.Current.WeatherObject[0].Main,
		"weatherObjectDescription": data.Current.WeatherObject[0].Description,
		"weatherObjectIcon":        data.Current.WeatherObject[0].Icon,
	}

	// executing the query string with the named arguments
	_, err := appConfig.DBPool.Exec(ctx, queryString, queryArguments)
	if err != nil {
		return fmt.Errorf("error in inserting open weather data into postgresql: %w", err)
	}

	return nil
}
