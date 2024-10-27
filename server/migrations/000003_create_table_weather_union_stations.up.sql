CREATE TABLE IF NOT EXISTS weather_union_stations(
    weather_station_id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    city_name TEXT NOT NULL,
    locality_name TEXT NOT NULL,
    locality_id TEXT NOT NULL,
    location geography(POINT, 4326) NOT NULL,
    device_type TEXT NOT NULL,
    device_type_integer INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);