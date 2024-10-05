package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/justinas/alice"
	"github.com/kelaaditya/zomato-weather-union/server/cmd/web/middlewares"
	"github.com/kelaaditya/zomato-weather-union/server/internal/config"
	"github.com/kelaaditya/zomato-weather-union/server/internal/models"
)

type application struct {
	logger      *slog.Logger
	config      *config.Config
	models      *models.Models
	middlewares *middlewares.Middleware
	// templateCache map[string]*template.Template
}

func main() {
	// create app level background context
	var ctx context.Context = context.Background()

	//
	// config
	//
	config := config.Config{}
	// initialize new configuration
	err := config.New(ctx)
	if err != nil {
		config.Logger.Error(err.Error())
		// force exit on error
		os.Exit(1)
	}
	// close the postgresql connection pool on function close
	defer config.DB.Close()

	// models
	models := models.Models{
		WeatherUnion:   &models.WeatherUnionModel{DB: config.DB},
		OpenWeatherMap: &models.OpenWeatherMapModel{DB: config.DB},
		Measurement:    &models.MeasurementModel{DB: config.DB},
		Calculation:    &models.CalculationModel{DB: config.DB},
	}

	// middlewares
	middlewares := middlewares.Middleware{
		Logger: config.Logger,
	}

	//
	// application
	//
	// create new application struct
	app := &application{
		config:      &config,
		models:      &models,
		middlewares: &middlewares,
	}

	//
	// HTTP mux
	//
	// create new HTTP multiplexer
	mux := http.NewServeMux()

	// compose chain starting with common headers and ending with recover panic
	// recover panic envelopes the entire system
	standard := alice.New(
		app.middlewares.RecoverPanic,
		app.middlewares.LogRequests,
		app.middlewares.CommonHeaders,
	)
	// link the routes handler to the middleware chain
	standard.Then(mux)

	// configuration for the http server
	// server := http.Server{
	// 	Addr: appConfig.ENVVariables.Port,
	// 	// Handler: ,
	// }
}
