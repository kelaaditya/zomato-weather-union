package main

import (
	"context"
	"os"
	"time"

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

	// ticker for periodic run of measurements fetching
	// and calculating temperatures
	// ticks are sent every hour
	ticker := time.NewTicker(time.Minute)
	// defer the graceful
	defer ticker.Stop()

	// goroutine to run measurement and calculation functions
	// on every tick from the ticker channel
	go func() {
		for t := range ticker.C {
			// server time is UTC
			// checking for tick at 05:00 UTC (10:30 IST)
			if t.Hour() == 5 && t.Minute() == 0 {
				// call APIs for measurement
				err = models.GetAndSaveMeasurementsFromAPISingleRun(
					appContext,
					&appConfig,
				)
				if err != nil {
					appConfig.Logger.Error(err.Error())
					os.Exit(1)
				}

				// calculate and save temperature values
				err = models.CalculateAndSaveTemperaturesAllUnprocessed(
					appContext,
					&appConfig,
				)
				if err != nil {
					appConfig.Logger.Error(err.Error())
					os.Exit(1)
				}
			}
		}
	}()

	// TODO
	// add HTTP server code here
}
