# zomato-weather-union
Code for the Weather Union (Zomato) project


## Database Structure
```postgresql
CREATE USER zomato_weather_union WITH PASSWORD '<PASSWORD>';
ALTER USER zomato_weather_union WITH SUPERUSER;
CREATE DATABASE zomato_weather_union;
GRANT ALL PRIVILEGES ON DATABASE zomato_weather_union TO zomato_weather_union;
```

```postgresql
CREATE EXTENSION postgis;
CREATE EXTENSION "uuid-ossp";
```

```postgresql
CREATE TABLE IF NOT EXISTS weather_union_stations(
    weather_station_id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    city_name TEXT NOT NULL,
    locality_name TEXT NOT NULL,
    locality_id TEXT NOT NULL,
    location geography(POINT, 4326) NOT NULL,
    device_type TEXT NOT NULL,
    device_type_integer INTEGER NOT NULL,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON weather_union_stations(locality_id);
```

```postgresql
CREATE TABLE IF NOT EXISTS temp_weather_union_stations(
    city_name TEXT NOT NULL,
    locality_name TEXT NOT NULL,
    locality_id TEXT NOT NULL,
    latitude FLOAT NOT NULL,
    longitude FLOAT NOT NULL,
    device_type TEXT NOT NULL,
    device_type_integer INTEGER NOT NULL
);
```

```postgresql
\COPY temp_weather_union_stations(
    city_name,
    locality_name,
    locality_id,
    latitude,
    longitude,
    device_type,
    device_type_integer
)
FROM '<PATH_TO_DATA_DIRECTORY>/weather-union-station-data.csv'
DELIMITER ','
CSV HEADER;
```

```postgresql
INSERT INTO weather_union_stations(
    city_name,
    locality_name,
    locality_id,
    location,
    device_type,
    device_type_integer
)
SELECT
    city_name,
    locality_name,
    locality_id,
    ST_Point(longitude, latitude),
    device_type,
    device_type_integer
FROM temp_weather_union_stations;
```

```postgresql
CREATE TABLE IF NOT EXISTS measurements_weather_union(
    measurement_id UUID PRIMARY KEY NOT NULL,
    weather_station_id UUID REFERENCES weather_union_stations(weather_station_id) NOT NULL,
    temperature FLOAT NOT NULL,
    humidity FLOAT NOT NULL,
    wind_speed FLOAT NOT NULL,
    wind_direction FLOAT NOT NULL,
    rain_intensity FLOAT NOT NULL,
    rain_accumulation FLOAT NOT NULL,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (measurement_id, weather_station_id)
);
```

```postgresql
CREATE TABLE IF NOT EXISTS measurements_open_weather_map(
    measurement_id UUID PRIMARY KEY NOT NULL REFERENCES measurements_weather_union(measurement_id),
    time_zone TEXT NOT NULL,
    time_zone_offset INTEGER NOT NULL,
    time_current BIGINT NOT NULL,
    time_sunrise BIGINT NOT NULL,
    time_sunset BIGINT NOT NULL,
    temperature FLOAT NOT NULL,
    feels_like FLOAT NOT NULL,
    pressure FLOAT NOT NULL,
    humidity FLOAT NOT NULL,
    dew_point FLOAT NOT NULL,
    uv_index FLOAT NOT NULL,
    clouds FLOAT NOT NULL,
    visibility BIGINT NOT NULL,
    wind_speed FLOAT NOT NULL,
    wind_direction FLOAT NOT NULL,
    wind_gust FLOAT NOT NULL,
    weather_object_id INTEGER NOT NULL,
    weather_object_main TEXT NOT NULL,
    weather_object_description TEXT NOT NULL,
    weather_object_icon TEXT NOT NULL,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
```