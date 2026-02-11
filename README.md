# IMIQ Project

This repository contains the infrastructure for the IMIQ project

## Run locally

Before starting the infrastructure, you must set a few environment variables:

| Variable | Description           | Example   |
|----------|-----------------------|-----------|
| API_KEY  | Key to access the api | secret123 |
| OPEN_WEATHER_MAP_API_KEY  | Key to open weather map (free account needed) | 1eas022db3a0qcaaa0299afbbaf47741 |
| DAILYPLAN_APP_SECRET | Secret for the dailyplan app | secret456 |
| WEATHER_STATION_PASSWORD | Password for Bresser9in1 Weather Station | SecretPw |
| THE_THINGS_NET_WEBHOOK_URL | Url for TTN Webhook | /data/secret_url |

You probably want to put these into an `.env` file:

```
API_KEY=secret123
OPEN_WEATHER_MAP_API_KEY=1eas022db3a0qcaaa0299afbbaf47741
DAILYPLAN_APP_SECRET=secret456
WEATHER_STATION_PASSWORD=SecretPw
THE_THINGS_NET_WEBHOOK_URL=/data/secret_url
```

Additionally, you should provide a `compose.override.yml` to specify at least your local port:

```yml
services:
  proxy:
    ports:
      - 8000:8000
```

Finally, you can start the infrastructure using `docker compose up -d`.

## Development

For development, you should mount source folders to the respective containers so that you can change the code without rebuilding the container.
You can achieve this by creating/editing a `compose.override.yml` file.
Please note, that you need to restart the stack after you have changed the `compose.override.yml` as follows:

```sh
docker compose down
docker compose up -d
```

The following sections explains local development for individual services with appropriate `compose.override.yml` files.

### Dashboard
Expand your `compose.override.yml` as follows:

```yml
services:
  dashboard:
    volumes:
      - ./services/dashboard:/app
    command: sleep infinity
```

Then jump into your container using `docker compose exec dashboard bash` and start development.
Then run:

```sh
npm run server
```

### Collector
Expand your `compose.override.yml` as follows:

```yml
services:
  collector:
    build:
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/collector:/app
    command: sleep infinity
```
Then jump into your container using `docker compose exec collector bash` and start development.
Then run:

```sh
go run . --config /app/config.yml
```

### Scheduled
Expand your `compose.override.yml` as follows:

```yml
services:
  scheduled:
    volumes:
      - ./services/scheduled:/app
    command: sleep infinity
```

Then jump into your container using `docker compose exec scheduled bash` and start development.
Then run:

```sh
python -u main.py
```

### Sensors
Expand your `compose.override.yml` as follows:

```yml
  sensors:
    build:
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/sensors/src:/app/src
    command: sleep infinity
```

Then jump into your container using `docker compose exec sensors bash` and start development.
Then run:

```sh
go run .
```

### Migrate
Expand your `compose.override.yml` as follows:

```yml
services:
  migrate:
    volumes:
      - ./services/migrate:/app
    command: sleep infinity
  crate:
    ports:
      - "4200:4200"   # crate web UI
```

Then jump into your container using `docker compose exec migrate bash` and start development.
Then run:

```sh
python -u main.py
```
