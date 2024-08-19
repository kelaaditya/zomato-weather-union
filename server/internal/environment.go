package internal

import (
	// external
	"os"

	"github.com/joho/godotenv"
)

type AppENVVariables struct {
	Port                 string
	DatabaseURL          string
	URLAPIWeatherUnion   string
	URLAPIOpenWeatherMap string
	APIKeyWeatherUnion   string
	APIKeyOpenWeatherMap string
}

// load environment variable values
// from .env to a struct and then return
func CreateEnvironmentStruct() (*AppENVVariables, error) {
	// load environment variables
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	// struct for storing values
	var envStruct AppENVVariables
	// get environment variables and
	// add them to the struct
	envStruct.Port = os.Getenv("PORT")
	envStruct.DatabaseURL = os.Getenv("DATABASE_URL")
	envStruct.URLAPIWeatherUnion = os.Getenv("URL_API_WEATHER_UNION_LOCALITY")
	envStruct.URLAPIOpenWeatherMap = os.Getenv("URL_API_OPEN_WEATHER_MAP")
	envStruct.APIKeyWeatherUnion = os.Getenv("API_KEY_WEATHER_UNION")
	envStruct.APIKeyOpenWeatherMap = os.Getenv("API_KEY_OPEN_WEATHER_MAP")

	return &envStruct, nil
}
