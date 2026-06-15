# Sherdog API

A small Go service that turns Sherdog fighter and event pages into JSON. It is built for experiments, prototypes, and local tooling where a lightweight HTTP interface is easier to work with than raw HTML.

The service scrapes on demand and can cache responses in Redis. Completed events return result data; upcoming events return the scheduled card with empty method, round, and time fields until Sherdog publishes results.

## Endpoints

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/fighter/:id` | Fighter profile, record summary, and fight history |
| `GET` | `/events/:id` | Event details and fight card |

The OpenAPI 3.1 schema lives in [`openapi.yaml`](openapi.yaml).

Sherdog slugs work as IDs:

```sh
curl http://localhost:8080/fighter/Jiri-Prochazka-97529
curl http://localhost:8080/events/UFC-295-Prochazka-vs-Pereira-98494
curl http://localhost:8080/events/UFC-Fight-Night-279-Kape-vs-Horiguchi-2-112139
```

Numeric fighter IDs are also accepted:

```sh
curl http://localhost:8080/fighter/97529
```

## Example Response

```json
{
  "name": "UFC Fight Night 279 - Kape vs. Horiguchi 2",
  "date": "Jun 20, 2026",
  "organization": "Ultimate Fighting Championship (UFC)",
  "fights": [
    {
      "fighters": [
        {
          "name": "Manel Kape",
          "url": "/fighter/Manel-Kape-101189",
          "outcome": "yet to come"
        },
        {
          "name": "Kyoji Horiguchi",
          "url": "/fighter/Kyoji-Horiguchi-64413",
          "outcome": "yet to come"
        }
      ],
      "way": "",
      "round": 0,
      "time": "",
      "weightclass": "Flyweight"
    }
  ]
}
```

## Run It

Docker Compose starts the API and Redis:

```sh
docker compose up --build
```

The API listens on `http://localhost:8080`.

For local development without Redis:

```sh
cd api-backend
CACHE_DISABLED=true go run ./cmd/sherdog-api
```

## Configuration

| Variable | Default | Notes |
| --- | --- | --- |
| `DB_CONN` | `redis://default:password@redis:6379` in Compose | Required when caching is enabled |
| `CACHE_DISABLED` | `false` | Set to `true`, `1`, `yes`, or `on` to scrape without Redis |

## Development

```sh
cd api-backend
go test ./...
```

## Project Layout

```text
api-backend/
  cmd/sherdog-api/          application entrypoint and service composition
  internal/domain/event/    event scraping, models, and cache decorator
  internal/domain/fighter/  fighter scraping, models, and cache decorator
  internal/database/redis/  Redis adapter for domain database interfaces
  internal/transport/http/  chi router and domain HTTP handlers
```

The scraper intentionally keeps the data model close to Sherdog's pages. Missing result values on upcoming cards are returned as zero values rather than guessed.
Fighter profiles include `upcomingfight`; it is an object when Sherdog lists a scheduled bout and `null` otherwise.

## Notes

This project is not affiliated with Sherdog. Be considerate with request volume, cache aggressively when you can, and treat the service as a proof of concept rather than a guaranteed data source.
