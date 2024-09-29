# zomato-weather-union
Code for the Weather Union (Zomato) project


## Deployment
### Installing Golang
Please follow the instructions to install the Go programming language here:
https://go.dev/doc/install

### PostgreSQL
Install PostgreSQL by running the following commands:
```bash
sudo apt install postgresql postgresql-contrib

# connect to postgres
sudo -u postgres psql
```

After connecting, set a new password for the Postgres user by running:
```sql
\password postgres
```

Once PostgreSQL is installed and running, create a database for the project and
its associated user by running the following commands inside PostgreSQL
```sql
CREATE DATABASE zomato_weather_union;
CREATE USER zomato_weather_union WITH ENCRYPTED PASSWORD '<PASSWORD>';

-- require for postgis
ALTER USER zomato_weather_union WITH SUPERUSER;

GRANT ALL PRIVILEGES ON DATABASE zomato_weather_union TO zomato_weather_union;

-- connect to the project database as the postgres user
\c zomato_weather_union postgres

-- grant all privileges on the public schema to the user zomato_weather_union
GRANT ALL PRIVILEGES ON SCHEMA public TO zomato_weather_union;

-- quit postgres
\q
```

Save the `zomato_weather_union` user password to the environment file.

### Database Migration
For database migrations, we will make use of the `golang-migrate` tool.
See: https://github.com/golang-migrate  
To install the tool, download the latest linux `amd64` tar file from their
release page here: https://github.com/golang-migrate/migrate/releases  
After downloading, untar by running the following command
```sh
tar -zxvf migrate.linux-amd64.tar.gz
```
After untaring, copy the `migrate` executable to the Golang binary folder
location.
Typically, it should be located at `/usr/local/go/bin`.
You might need the super-user permission.
Run the following command:
```sh
mv migrate /usr/local/go/bin/migrate
```

To carry out the database migration, run the following command:
```sh
migrate --path=server/migrations --database=$DATABASE_URL up
```

Note that the shell variable `DATABASE_URL` has been loaded into the
current session.
This can be done by loading the environment file into the shell.
Else, manually enter the database connection URL value into this command.
The database connection URL format is: https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS

### Add Weather Station Data To Database
Copy the weather station data CSV file to `/tmp`.
```sh
cp data/weather-union-station-data.csv /tmp/
```

```sh
psql $DATABASE_URL -f scripts/copy-weather-station-data-into-database.sql
```