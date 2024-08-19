package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/kelaaditya/zomato-weather-union/server/internal"
)

func main() {
	// create application level context
	var appContext context.Context = context.Background()

	// application config struct
	var appConfig internal.AppConfig

	// logger
	appConfig.Logger = internal.CreateLogger()

	// environment variables
	appENVStruct, err := internal.CreateEnvironmentStruct()
	if err != nil {
		appConfig.Logger.Error(err.Error())
	}
	appConfig.ENVVariables = appENVStruct

	// create database pool and store into app config
	appDBPool, err := internal.CreateDBPool(&appContext, appConfig.ENVVariables.DatabaseURL)
	if err != nil {
		appConfig.Logger.Error(err.Error())
		// logger does not auto-exit
		// manual call required
		os.Exit(1)
	}
	appConfig.DBPool = appDBPool

	// using the plain http mux
	var mux *http.ServeMux = http.NewServeMux()

	// get port environment variable
	var portHTTP string = appConfig.ENVVariables.Port

	// http server configuration
	var server *http.Server = &http.Server{
		Addr:         ":" + portHTTP,
		Handler:      mux,
		ErrorLog:     slog.NewLogLogger(appConfig.Logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	// print if PORT variable is defined
	appConfig.Logger.Info("starting the http server", "port", portHTTP)

	// create a http server
	err = server.ListenAndServe()
	if err != nil {
		appConfig.Logger.Error(err.Error())
		// logger does not auto-exit
		// manual call required
		os.Exit(1)
	}
}
