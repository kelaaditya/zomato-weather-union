package config

import (
	// external
	"os"

	"github.com/joho/godotenv"
)

type Environment struct {
	Port                    string
	DatabaseURL             string
	URLBaseWeatherUnion     string
	URLBaseOpenWeatherMap   string
	APIKeyWeatherUnion      string
	APIKeyOpenWeatherMap    string
	PathToPythonEnvironment string
}

// load environment variable values
// from .env to a struct and then return
func (config *Config) initializeEnvironment() error {
	// load environment variables
	err := godotenv.Load()
	if err != nil {
		return err
	}

	// struct for storing values
	var newEnvironment Environment
	// get environment variables and
	// add them to the struct
	newEnvironment.Port = os.Getenv("PORT")
	newEnvironment.DatabaseURL = os.Getenv("DATABASE_URL")
	newEnvironment.URLBaseWeatherUnion = os.Getenv("URL_BASE_WEATHER_UNION")
	newEnvironment.URLBaseOpenWeatherMap = os.Getenv("URL_BASE_OPEN_WEATHER_MAP")
	newEnvironment.APIKeyWeatherUnion = os.Getenv("API_KEY_WEATHER_UNION")
	newEnvironment.APIKeyOpenWeatherMap = os.Getenv("API_KEY_OPEN_WEATHER_MAP")
	newEnvironment.PathToPythonEnvironment = os.Getenv("PATH_TO_PYTHON_ENVIRONMENT")

	// configured environment variables struct
	config.Environment = &newEnvironment

	// return nil if all okay
	return nil
}
