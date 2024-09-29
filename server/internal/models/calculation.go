package models

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kelaaditya/zomato-weather-union/server/internal"
	"golang.org/x/sync/errgroup"
)

// the "calculation" type contains all parameters
// that need to be calculated in one iteration of the API call
type Calculation struct {
	DewPointTemperature float64 `json:"dew_point_temperature"`
	WetBulbTemperature  float64 `json:"wet_bulb_temperature"`
}

// get all calculations across all weather stations
func GetAllCalculations(
	ctx context.Context,
	appConfig *internal.AppConfig,
) error {
	// initialize runID as UUID for this run
	runID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	// get data of all weather stations
	stations, err := GetWeatherStationDataFromWeatherUnion(ctx, appConfig)
	if err != nil {
		return err
	}
	fmt.Println(stations)

	// create a wait group
	var wgCalculations errgroup.Group

	// iterate over the slice of station data
	for _, station := range stations {
		wgCalculations.Go(func() error {
			// run one unit of calculation
			err := GetOneUnitOfCalculation(
				ctx,
				appConfig,
				runID,
				station,
			)
			if err != nil {
				return err
			}

			return nil
		})
	}

	// wait until all goroutines are completed
	// return first non-nil error
	if err := wgCalculations.Wait(); err != nil {
		return err
	}

	return nil
}

// complete one unit of calculation of the python script
func GetOneUnitOfCalculation(
	ctx context.Context,
	appConfig *internal.AppConfig,
	runID uuid.UUID,
	station WeatherUnionDataStation,
) error {
	// initialize mesurementID as UUID
	measurementID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	// get data from weather union
	dataWeatherUnion, err := GetWeatherDataFromWeatherUnionLocality(
		appConfig,
		station.LocalityID,
	)
	if err != nil {
		return err
	}

	// check if station is unavailable
	if dataWeatherUnion.Message == nil ||
		(dataWeatherUnion.Message != nil &&
			*dataWeatherUnion.Message == "temporarily unavailable") {
		return fmt.Errorf("station with ID %v is unavailable", station.WeatherStationID)
	}

	// if rain guage (type 2)
	// if yes, return error
	if dataWeatherUnion.LocalityWeatherData.Temperature == nil ||
		dataWeatherUnion.LocalityWeatherData.Humidity == nil ||
		dataWeatherUnion.LocalityWeatherData.WindSpeed == nil ||
		dataWeatherUnion.LocalityWeatherData.WindDirection == nil {
		return fmt.Errorf("station with ID %v was rain gauge", station.WeatherStationID)
	}

	// get data from open weather map
	dataOpenWeatherMap, err := GetWeatherDataFromOpenWeatherMap(
		appConfig,
		station.Latitude,
		station.Longitude,
	)
	if err != nil {
		return err
	}
	fmt.Println(station)

	// calculate the required parameters using the python script
	calculation, err := CallPythonScriptForOneUnitOfCalculation(
		*dataWeatherUnion.LocalityWeatherData.Temperature,
		*dataWeatherUnion.LocalityWeatherData.Humidity,
		*dataOpenWeatherMap.Current.Pressure,
	)
	if err != nil {
		fmt.Println("Error here!")
		return err
	}

	// save data from weather union
	err = SaveWeatherDataFromWeatherUnion(
		ctx,
		appConfig,
		measurementID,
		station.WeatherStationID,
		runID,
		&dataWeatherUnion,
	)
	if err != nil {
		fmt.Println("Error here 2!")
		return err
	}

	fmt.Println(dataOpenWeatherMap.Current.Pressure)
	// save data from open weather map
	err = SaveWeatherDataFromOpenWeatherMap(
		ctx,
		appConfig,
		measurementID,
		station.WeatherStationID,
		runID,
		&dataOpenWeatherMap,
	)
	if err != nil {
		fmt.Println("Error here 3!")
		return err
	}

	// save the calculation data
	err = SaveCalculation(
		ctx,
		appConfig,
		measurementID,
		calculation,
	)
	if err != nil {
		return err
	}

	// return nil if all okay
	return nil
}

// run the python script from the fetched
// data of one instance
func CallPythonScriptForOneUnitOfCalculation(
	temperature float64,
	humidity float64,
	pressure float64,
) (Calculation, error) {
	// placeholder for calculation struct
	var calculation Calculation

	// build command line argument strings
	var CLATemperature = "--temperature=" + fmt.Sprint(temperature)
	var CLAHumidity = "--humidity=" + fmt.Sprint(humidity)
	var CLAPressure = "--pressure=" + fmt.Sprint(pressure)

	// create os command
	// to debug the command, do this
	// fmt.Println(command.String())
	command := exec.Command(
		"python3",
		"../scripts/wet_bulb_temperature.py",
		CLATemperature,
		CLAHumidity,
		CLAPressure,
	)

	// get the stdout after running the command
	stdout, err := command.Output()
	if err != nil {
		return Calculation{}, err
	}

	// convert the JSON slice of bytes to struct
	err = json.Unmarshal(stdout, &calculation)
	if err != nil {
		return Calculation{}, err
	}

	return calculation, nil
}

// save the calculation struct parameters to the database
func SaveCalculation(
	ctx context.Context,
	appConfig *internal.AppConfig,
	measurementID uuid.UUID,
	calculation Calculation,
) error {
	// postgresql query string
	var queryString string = `
	INSERT INTO calculations(
		measurement_id,
		method,
		dew_point_temperature,
		wet_bulb_temperature
	)
	VALUES (
		@measurementID,
		@method,
		@dewPointTemperature,
		@wetBulbTemperature
	);
	`

	// named arguments for building the query string
	var queryArguments pgx.NamedArgs = pgx.NamedArgs{
		"measurementID":       measurementID,
		"method":              "metpy-with-open-weather-map",
		"dewPointTemperature": calculation.DewPointTemperature,
		"wetBulbTemperature":  calculation.WetBulbTemperature,
	}

	// executing the query string with the named arguments
	_, err := appConfig.DBPool.Exec(ctx, queryString, queryArguments)
	if err != nil {
		return fmt.Errorf("error in inserting calculation into postgresql: %w", err)
	}

	return nil
}
