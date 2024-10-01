CREATE TABLE IF NOT EXISTS measurements_open_weather_map(
    measurement_id UUID PRIMARY KEY NOT NULL,
    weather_station_id UUID NOT NULL REFERENCES weather_union_stations(weather_station_id),
    run_id UUID NOT NULL REFERENCES measurement_runs(run_id),
    time_zone TEXT,
    time_zone_offset INTEGER,
    time_current BIGINT,
    time_sunrise BIGINT,
    time_sunset BIGINT,
    temperature FLOAT,
    feels_like FLOAT,
    pressure FLOAT,
    humidity FLOAT,
    dew_point FLOAT,
    uv_index FLOAT,
    clouds FLOAT,
    visibility BIGINT,
    wind_speed FLOAT,
    wind_direction FLOAT,
    wind_gust FLOAT,
    weather_object_id INTEGER,
    weather_object_main TEXT,
    weather_object_description TEXT,
    weather_object_icon TEXT,
    is_processed_for_calculation_temperature BOOLEAN NOT NULL DEFAULT FALSE,
    is_successful_for_calculation_temperature BOOLEAN NOT NULL DEFAULT FALSE,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (weather_station_id, run_id)
);