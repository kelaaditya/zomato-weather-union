package main

import (
	"context"

	"github.com/kelaaditya/zomato-weather-union/server/internal"
	"github.com/kelaaditya/zomato-weather-union/server/internal/models"
)

func main() {
	// create application level context
	var appContext context.Context = context.Background()

	// application config struct
	var appConfig internal.AppConfig

	// initialize logger
	appConfig.LoggerInitialize()

	// initialize environment variables
	err := appConfig.ENVInitialize()
	if err != nil {
		appConfig.Logger.Error(err.Error())
	}

	// initialize database connection pool
	err = appConfig.DBPoolInitialize(
		appContext,
		appConfig.ENVVariables.DatabaseURL,
	)
	if err != nil {
		appConfig.Logger.Error(err.Error())
	}
	defer appConfig.DBPoolClose()

	// call APIs for measurement
	err = models.GetAndSaveMeasurementsFromAPISingleRun(
		appContext,
		&appConfig,
	)
	if err != nil {
		appConfig.Logger.Error(err.Error())
	}

	// calculate and save temperature values
	err = models.CalculateAndSaveTemperaturesAllUnprocessed(
		appContext,
		&appConfig,
	)
	if err != nil {
		appConfig.Logger.Error(err.Error())
	}
}
