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

// structure of weather reponse from open weather map
type OpenWeatherAPIReponse struct {
	Latitude       *float64                      `json:"lat"`
	Longitude      *float64                      `json:"lon"`
	TimeZone       *string                       `json:"timezone"`
	TimeZoneOffset *int64                        `json:"timezone_offset"`
	Current        OpenWeatherDataWeatherCurrent `json:"current"`
}

type OpenWeatherDataWeatherCurrent struct {
	TimeCurrent   *int64                        `json:"dt"`
	TimeSunrise   *int64                        `json:"sunrise"`
	TimeSunset    *int64                        `json:"sunset"`
	Temperature   *float64                      `json:"temp"`
	FeelsLike     *float64                      `json:"feels_like"`
	Pressure      *float64                      `json:"pressure"`
	Humidity      *float64                      `json:"humidity"`
	DewPoint      *float64                      `json:"dew_point"`
	UVIndex       *float64                      `json:"uvi"`
	Clouds        *float64                      `json:"clouds"`
	Visibility    *int64                        `json:"visibility"`
	WindSpeed     *float64                      `json:"wind_speed"`
	WindDirection *float64                      `json:"wind_deg"`
	WindGust      *float64                      `json:"wind_gust"`
	WeatherObject []OpenWeatherSubWeatherObject `json:"weather"`
}

type OpenWeatherSubWeatherObject struct {
	ID          *int    `json:"id"`
	Main        *string `json:"main"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}

// function for calling the Open Weather API
func GetWeatherDataFromOpenWeatherMap(
	appConfig *internal.AppConfig,
	latitude string,
	longitude string,
) (OpenWeatherAPIReponse, error) {
	// base URL
	var baseURLString string = appConfig.ENVVariables.URLBaseOpenWeatherMap
	var oneCallPathString string = "onecall"
	// join base URL with specific path
	specificURLString, err := url.JoinPath(baseURLString, oneCallPathString)
	if err != nil {
		return OpenWeatherAPIReponse{}, err
	}
	// parse to URL type
	specificURL, err := url.Parse(specificURLString)
	if err != nil {
		return OpenWeatherAPIReponse{}, err
	}
	// add locality ID as the query parameter
	q := specificURL.Query()
	q.Add("lat", latitude)
	q.Add("lon", longitude)
	q.Add("exclude", "minutely,hourly,daily,alerts")
	q.Add("appid", appConfig.ENVVariables.APIKeyOpenWeatherMap)
	specificURL.RawQuery = q.Encode()
	// get string of the built API URL
	var APIURLString string = specificURL.String()
	// carry out get request
	response, err := http.Get(APIURLString)
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
func SaveWeatherDataFromOpenWeatherMap(
	ctx context.Context,
	appConfig *internal.AppConfig,
	measurementID uuid.UUID,
	weatherStationID uuid.UUID,
	runID uuid.UUID,
	data *OpenWeatherAPIReponse,
) error {
	// postgresql query string
	var queryString string = `
	INSERT INTO measurements_open_weather_map(
		measurement_id,
		weather_station_id,
		run_id,
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
		weather_object_icon,
		is_processed_for_wet_bulb_calculation
	)
	VALUES (
		@measurementID,
		@weatherStationID,
		@runID,
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
		@weatherObjectIcon,
		@isProcessedForWetBulbCalculation
	);
	`

	// named arguments for building the query string
	var queryArguments pgx.NamedArgs = pgx.NamedArgs{
		"measurementID":                    measurementID,
		"weatherStationID":                 weatherStationID,
		"runID":                            runID,
		"timeZone":                         utilities.DereferenceOrNil(data.TimeZone),
		"timeZoneOffset":                   utilities.DereferenceOrNil(data.TimeZoneOffset),
		"timeCurrent":                      utilities.DereferenceOrNil(data.Current.TimeCurrent),
		"timeSunrise":                      utilities.DereferenceOrNil(data.Current.TimeSunrise),
		"timeSunset":                       utilities.DereferenceOrNil(data.Current.TimeSunset),
		"temperature":                      utilities.DereferenceOrNil(data.Current.Temperature),
		"feelsLike":                        utilities.DereferenceOrNil(data.Current.FeelsLike),
		"pressure":                         utilities.DereferenceOrNil(data.Current.Pressure),
		"humidity":                         utilities.DereferenceOrNil(data.Current.Humidity),
		"dewPoint":                         utilities.DereferenceOrNil(data.Current.DewPoint),
		"UVIndex":                          utilities.DereferenceOrNil(data.Current.UVIndex),
		"clouds":                           utilities.DereferenceOrNil(data.Current.Clouds),
		"visibility":                       utilities.DereferenceOrNil(data.Current.Visibility),
		"windSpeed":                        utilities.DereferenceOrNil(data.Current.WindSpeed),
		"windDirection":                    utilities.DereferenceOrNil(data.Current.WindDirection),
		"windGust":                         utilities.DereferenceOrNil(data.Current.WindGust),
		"weatherObjectID":                  utilities.DereferenceOrNil(data.Current.WeatherObject[0].ID),
		"weatherObjectMain":                utilities.DereferenceOrNil(data.Current.WeatherObject[0].Main),
		"weatherObjectDescription":         utilities.DereferenceOrNil(data.Current.WeatherObject[0].Description),
		"weatherObjectIcon":                utilities.DereferenceOrNil(data.Current.WeatherObject[0].Icon),
		"isProcessedForWetBulbCalculation": false, // set false manually as this is just the data entry step
	}

	// executing the query string with the named arguments
	_, err := appConfig.DBPool.Exec(ctx, queryString, queryArguments)
	if err != nil {
		return fmt.Errorf(
			"error in inserting open weather data into postgresql: %w",
			err,
		)
	}

	return nil
}
