package models

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kelaaditya/zomato-weather-union/server/internal"
)

// the "calculation" type contains all parameters
// that need to be calculated in one iteration of the API call
type Calculation struct {
	DewPointTemperature float64 `json:"dew_point_temperature"`
	WetBulbTemperature  float64 `json:"wet_bulb_temperature"`
}

func CalculateWetBulbTemperature() error {
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
		return fmt.Errorf(
			"error in inserting calculation into postgresql: %w",
			err,
		)
	}

	return nil
}
