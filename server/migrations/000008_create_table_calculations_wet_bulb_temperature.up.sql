CREATE TABLE IF NOT EXISTS calculations_wet_bulb_temperature(
    calculation_id UUID PRIMARY KEY NOT NULL,
    measurement_id_weather_union UUID NOT NULL REFERENCES measurements_weather_union(measurement_id),
    method TEXT NOT NULL CHECK (method IN (
        'metpy-with-open-weather-map'
    )),
    dew_point_temperature FLOAT NOT NULL,
    wet_bulb_temperature FLOAT NOT NULL,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);