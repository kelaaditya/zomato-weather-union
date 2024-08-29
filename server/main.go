package main

import (
	"context"
	"os"

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
	err = appConfig.DBPoolInitialize(appContext, appConfig.ENVVariables.DatabaseURL)
	if err != nil {
		appConfig.Logger.Error(err.Error())
		// logger does not auto-exit
		// manual call required
		os.Exit(1)
	}
	defer appConfig.DBPoolClose()

	// do one pass through all weather stations
	if err := models.GetAllCalculations(appContext, &appConfig); err != nil {
		appConfig.Logger.Error(err.Error())
		// manual call exit
		os.Exit(1)
	}
}
