package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kelaaditya/zomato-weather-union/server/internal"
)

// structure of weather reponse from open weather map
type OpenWeatherAPIReponse struct {
	Latitude       float64                       `json:"lat"`
	Longitude      float64                       `json:"lon"`
	Timezone       string                        `json:"timezone"`
	TimezoneOffset int64                         `json:"timezone_offset"`
	Current        OpenWeatherCurrentWeatherData `json:"current"`
}

type OpenWeatherCurrentWeatherData struct {
	CurrentTime   int64                         `json:"dt"`
	SunriseTime   int64                         `json:"sunrise"`
	SunsetTime    int64                         `json:"sunset"`
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
