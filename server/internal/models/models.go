package models

type Models struct {
	WeatherUnion   *WeatherUnionModel
	OpenWeatherMap *OpenWeatherMapModel
	Measurement    *MeasurementModel
	Calculation    *CalculationModel
}
