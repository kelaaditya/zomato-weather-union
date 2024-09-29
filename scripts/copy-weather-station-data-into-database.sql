/*
 * script to copy the Weather Union station data into the database
 */


/*
 * create temporary table for storing weather station data from weather union
 * (directly from the CSV file) so that latitude, longitude manipulation can
 * be done into the geography type of Postgis
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
FROM '/tmp/weather-union-station-data.csv'
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