CREATE TABLE IF NOT EXISTS calculations_temperature(
    calculation_id UUID PRIMARY KEY NOT NULL,
    measurement_id_weather_union UUID NOT NULL REFERENCES measurements_weather_union(measurement_id),
    method TEXT NOT NULL CHECK (method IN (
        'metpy-with-open-weather-map'
    )),
    temperature_dew_point FLOAT NOT NULL,
    temperature_wet_bulb FLOAT NOT NULL,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);