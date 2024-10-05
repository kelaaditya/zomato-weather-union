package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/justinas/alice"
	"github.com/kelaaditya/zomato-weather-union/server/cmd/web/handlers"
	"github.com/kelaaditya/zomato-weather-union/server/cmd/web/middlewares"
	"github.com/kelaaditya/zomato-weather-union/server/internal/config"
	"github.com/kelaaditya/zomato-weather-union/server/internal/models"
	"github.com/kelaaditya/zomato-weather-union/server/ui"
)

// application level configurations and operations
type application struct {
	config      *config.Config
	handlers    *handlers.Handler
	middlewares *middlewares.Middleware
}

func main() {
	//
	// application
	//
	var app application

	//
	// context
	//
	// create app level background context
	var ctx context.Context = context.Background()

	//
	// config
	//
	app.config = &config.Config{}
	// initialize the configurations of the logger, environment and
	// database
	err := app.config.New(ctx)
	if err != nil {
		app.config.Logger.Error(err.Error())
		// force exit on error
		os.Exit(1)
	}
	// close the postgresql connection pool on function close
	defer app.config.DB.Close()

	//
	// models
	//
	models := &models.Models{
		WeatherUnion:   &models.WeatherUnionModel{DB: app.config.DB},
		OpenWeatherMap: &models.OpenWeatherMapModel{DB: app.config.DB},
		Measurement:    &models.MeasurementModel{DB: app.config.DB},
		Calculation:    &models.CalculationModel{DB: app.config.DB},
	}

	//
	// html template cache
	//
	HTMLTemplateCache, err := ui.CreateHTMLTemplateCache()
	if err != nil {
		app.config.Logger.Error(err.Error())
		// force exit on error
		os.Exit(1)
	}

	//
	// handlers
	//
	app.handlers = &handlers.Handler{
		Logger:        app.config.Logger,
		TemplateCache: HTMLTemplateCache,
		Models:        models,
	}

	//
	// middlewares
	//
	app.middlewares = &middlewares.Middleware{
		Logger: app.config.Logger,
	}

	//
	// http mux
	//
	// create new HTTP multiplexer
	mux := http.NewServeMux()
	// handlers
	// static file server for the local file system
	var fileServer http.Handler = http.FileServer(http.Dir("./ui/static/"))
	// handle req
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	// attaching the home handler to the mux
	// restrict subtree paths using `${1}`
	mux.HandleFunc("GET /{$}", app.handlers.Home())

	// compose chain starting with common headers and ending with recover panic
	// recover panic envelopes the entire system
	// link the routes handler to the middleware chain
	muxWithMiddleware := alice.New(
		app.middlewares.RecoverPanic,
		app.middlewares.LogRequests,
		app.middlewares.CommonHeaders,
	).Then(mux)

	//
	// http server
	//
	// port
	var HTTPPort string = app.config.Environment.Port
	// server config
	server := http.Server{
		Addr:    ":" + HTTPPort,
		Handler: muxWithMiddleware,
		ErrorLog: slog.NewLogLogger(
			app.config.Logger.Handler(),
			slog.LevelError,
		),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	// print if port variable is fetched
	app.config.Logger.Info(
		"starting a http server",
		"port",
		HTTPPort,
	)

	// create a http server
	err = server.ListenAndServe()
	if err != nil {
		app.config.Logger.Error(err.Error())
		// force exit on error
		os.Exit(1)
	}
}
