/***************************************
 * script to initialize the database
 * and the tables for the project
 **************************************/

/* 
 * create database and user
 */
CREATE USER zomato_weather_union WITH PASSWORD :v1; -- variable for password
ALTER USER zomato_weather_union WITH SUPERUSER;
CREATE DATABASE zomato_weather_union;
GRANT ALL PRIVILEGES ON DATABASE zomato_weather_union TO zomato_weather_union;

/*
 * connect to the weather union database
 */
\c zomato_weather_union

/*
 * add extensions to the weather union database
 */
CREATE EXTENSION postgis;
CREATE EXTENSION "uuid-ossp";

/*
 * create table for storing weather station data from weather data
 */
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

/*
 * create temporary table for storing weather station data from
 * weather union (directly from the CSV file) so that latitude,
 * longitude manipulation can be done into the geography type of
 * Postgis
 */
CREATE TABLE IF NOT EXISTS temp_weather_union_stations(
    city_name TEXT NOT NULL,
    locality_name TEXT NOT NULL,
    locality_id TEXT NOT NULL,
    latitude FLOAT NOT NULL,
    longitude FLOAT NOT NULL,
    device_type TEXT NOT NULL,
    device_type_integer INTEGER NOT NULL
);

/*
 * import CSV data into the temporary table
 */
COPY temp_weather_union_stations(
    city_name,
    locality_name,
    locality_id,
    latitude,
    longitude,
    device_type,
    device_type_integer
)
FROM :v2
DELIMITER ','
CSV HEADER;

/*
 * insert data from temp to actual weather station table
 */
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

/*
 * drop the temporary table
 */
DROP TABLE temp_weather_union_stations;

/*
 * create table for the weather union measurements
 */
CREATE TABLE IF NOT EXISTS measurements_weather_union(
    measurement_id UUID PRIMARY KEY NOT NULL,
    run_id UUID NOT NULL,
    weather_station_id UUID REFERENCES weather_union_stations(weather_station_id) NOT NULL,
    message TEXT NOT NULL,
    device_type INTEGER NOT NULL,
    temperature FLOAT NOT NULL,
    humidity FLOAT NOT NULL,
    wind_speed FLOAT NOT NULL,
    wind_direction FLOAT NOT NULL,
    rain_intensity FLOAT NOT NULL,
    rain_accumulation FLOAT NOT NULL,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (measurement_id, weather_station_id)
);

/*
 * create table for the open weather map measurements
 */
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
);

/*
 * create table to store
 */
CREATE TABLE IF NOT EXISTS calculations(
    measurement_id UUID PRIMARY KEY NOT NULL REFERENCES measurements_weather_union(measurement_id),
    method TEXT NOT NULL CHECK (method IN (
        'metpy-with-open-weather-map'
    )),
    dew_point_temperature FLOAT NOT NULL,
    wet_bulb_temperature FLOAT NOT NULL,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);