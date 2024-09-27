# zomato-weather-union
Code for the Weather Union (Zomato) project


## Deployment
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
GRANT ALL PRIVILEGES ON DATABASE zomato_weather_union TO zomato_weather_union;

-- connect to the project database as the postgres user
\c zomato_weather_union postgres

-- grant all privileges on the public schema to the user zomato_weather_union
GRANT ALL PRIVILEGES ON SCHEMA public TO zomato_weather_union;

-- quit postgres
\q
```