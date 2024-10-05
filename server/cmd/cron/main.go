package main

import (
	"context"
	"os"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/kelaaditya/zomato-weather-union/server/internal/config"
	"github.com/kelaaditya/zomato-weather-union/server/internal/models"
	"golang.org/x/sync/errgroup"
)

// application level configurations and operations
type application struct {
	config *config.Config
	models *models.Models
}

func main() {
	//
	var app application

	// create app level background context
	var ctx context.Context = context.Background()

	//
	// config
	//
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
	app.models = &models.Models{
		WeatherUnion:   &models.WeatherUnionModel{DB: app.config.DB},
		OpenWeatherMap: &models.OpenWeatherMapModel{DB: app.config.DB},
		Measurement:    &models.MeasurementModel{DB: app.config.DB},
		Calculation:    &models.CalculationModel{DB: app.config.DB},
	}

	err = app.GetAndSaveMeasurementsFromAPISingleRun(ctx)
	if err != nil {
		app.config.Logger.Error(err.Error())
		// force exit on error
		os.Exit(1)
	}

	err = app.CalculateAndSaveTemperaturesAllUnprocessed(ctx)
	if err != nil {
		app.config.Logger.Error(err.Error())
		// force exit on error
		os.Exit(1)
	}
}

// carry out a single run of measurements over all the
// weather stations from weather union
func (app *application) GetAndSaveMeasurementsFromAPISingleRun(
	ctx context.Context,
) error {
	// initialize runID as UUID for this calculation
	runID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	// log runID when started
	app.config.Logger.Info("run started.", "runID", runID.String())

	// get all weather station data from weather union
	sliceStationsWeatherUnion, err :=
		app.models.WeatherUnion.GetWeatherStationsAllWeatherUnion(
			ctx,
		)
	if err != nil {
		return err
	}

	// create a slice to append measurements from weather union
	var sliceMeasurementsWeatherUnion []models.WeatherUnionMeasurement
	// create a slice to append measurements from weather union
	var sliceMeasurementsOpenWeatherMap []models.OpenWeatherMapMeasurement

	// create a wait group
	var wgMeasurements errgroup.Group
	// create a mutex object
	var mutex sync.Mutex

	// iterate over all stations
	for _, station := range sliceStationsWeatherUnion {
		wgMeasurements.Go(func() error {
			// carry out API call to weather union
			measurementWeatherUnion, err :=
				app.models.WeatherUnion.CallAPIWeatherUnionLocality(
					app.config.Environment.URLBaseWeatherUnion,
					app.config.Environment.APIKeyWeatherUnion,
					&station,
					runID,
				)
			if err != nil {
				return err
			}

			// carry out API call to open weather map
			measurementOpenWeatherMap, err :=
				app.models.OpenWeatherMap.CallAPIOpenWeatherMap(
					app.config.Environment.URLBaseOpenWeatherMap,
					app.config.Environment.APIKeyOpenWeatherMap,
					&station,
					runID,
				)
			if err != nil {
				return err
			}

			// append new measurements to corresponding slices
			// lock and unlock slices while appending
			mutex.Lock()
			sliceMeasurementsWeatherUnion = append(
				sliceMeasurementsWeatherUnion,
				measurementWeatherUnion,
			)
			sliceMeasurementsOpenWeatherMap = append(
				sliceMeasurementsOpenWeatherMap,
				measurementOpenWeatherMap,
			)
			mutex.Unlock()

			// return nil if okay
			return nil
		})
	}

	// wait until all goroutines are completed
	// return the first non-nil error
	if err := wgMeasurements.Wait(); err != nil {
		// do not return an error here
		// as we want to continue the flag setting for those that
		// have been processed
		app.config.Logger.Error(err.Error())
	}

	// log the count of measurements received from weather union
	app.config.Logger.Info(
		"measurements gathered from weather union",
		"total",
		strconv.Itoa(len(sliceMeasurementsWeatherUnion)),
	)
	// log the count of measurements received from open weather map
	app.config.Logger.Info(
		"measurements gathered from open weather map",
		"total",
		strconv.Itoa(len(sliceMeasurementsOpenWeatherMap)),
	)

	// save run ID
	err = app.models.Measurement.SaveMeasurementRun(ctx, runID)
	if err != nil {
		return err
	}

	// save measurements from weather union
	err = app.models.WeatherUnion.SaveMeasurementsWeatherUnion(
		ctx,
		sliceMeasurementsWeatherUnion,
	)
	if err != nil {
		return err
	}

	// save measurement from open weather map
	err = app.models.OpenWeatherMap.SaveMeasurementOpenWeatherMap(
		ctx,
		sliceMeasurementsOpenWeatherMap,
	)
	if err != nil {
		return err
	}

	// return nil if all okay
	return nil
}

// calculate all unprocessed wet bulb temperature values
// get the unprocessed data and then return a slice containing the values
func (app *application) CalculateAndSaveTemperaturesAllUnprocessed(
	ctx context.Context,
) error {
	// get all unprocessed measurements
	sliceMeasurementsUnprocessed, err :=
		app.models.Measurement.GetUnprocessedDataForCalculationsTemperature(
			ctx,
		)
	if err != nil {
		return err
	}

	// create a slice to append calculations to
	var sliceCalculationsSuccessful []models.CalculationTemperature

	// create a wait group
	var wgCalculations errgroup.Group
	// create a mutex object
	var mutex sync.Mutex

	// iterate over measurements
	for _, measurement := range sliceMeasurementsUnprocessed {
		wgCalculations.Go(func() error {
			// carry out calculations over a single measurement
			calculation, err :=
				app.models.Calculation.CalculateTemperatureFromSingleMeasurement(
					measurement,
				)
			if err != nil {
				return err
			}

			// append new successful calculation to slice of all successful
			// calculations
			// lock and unlock slice while appending
			mutex.Lock()
			sliceCalculationsSuccessful = append(
				sliceCalculationsSuccessful,
				calculation,
			)
			mutex.Unlock()

			// return nil if okay
			return nil
		})
	}

	// wait until all goroutines are completed
	// return the first non-nil error
	if err := wgCalculations.Wait(); err != nil {
		// do not return an error here
		// as we want to continue the flag setting for those that
		// have been processed
		app.config.Logger.Error(err.Error())
	}

	// save calculations
	err = app.models.Calculation.SaveCalculationsTemperatures(
		ctx,
		sliceCalculationsSuccessful,
	)
	if err != nil {
		return err
	}

	// set flag is_processed for weather union measurements
	err = app.models.Calculation.SetFlagsTemperature(
		ctx,
		sliceMeasurementsUnprocessed,
		sliceCalculationsSuccessful,
	)
	if err != nil {
		return err
	}

	// return nil if all okay
	return nil
}
