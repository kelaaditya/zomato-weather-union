CREATE TABLE IF NOT EXISTS measurements_weather_union(
    measurement_id UUID PRIMARY KEY NOT NULL,
    weather_station_id UUID NOT NULL REFERENCES weather_union_stations(weather_station_id),
    run_id UUID NOT NULL REFERENCES measurement_runs(run_id),
    message TEXT NOT NULL,
    device_type INTEGER NOT NULL,
    temperature FLOAT NOT NULL,
    humidity FLOAT NOT NULL,
    wind_speed FLOAT NOT NULL,
    wind_direction FLOAT NOT NULL,
    rain_intensity FLOAT NOT NULL,
    rain_accumulation FLOAT NOT NULL,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (measurement_id, weather_station_id),
    UNIQUE (weather_station_id, run_id)
);