CREATE TABLE IF NOT EXISTS measurements_weather_union(
    measurement_id UUID PRIMARY KEY NOT NULL,
    weather_station_id UUID NOT NULL REFERENCES weather_union_stations(weather_station_id),
    run_id UUID NOT NULL REFERENCES measurement_runs(run_id),
    message TEXT,
    device_type INTEGER,
    temperature FLOAT,
    humidity FLOAT,
    wind_speed FLOAT,
    wind_direction FLOAT,
    rain_intensity FLOAT,
    rain_accumulation FLOAT,
    is_processed_for_wet_bulb_calculation BOOLEAN NOT NULL DEFAULT FALSE,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (measurement_id, weather_station_id),
    UNIQUE (weather_station_id, run_id)
);