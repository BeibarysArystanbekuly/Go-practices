# Polling / Voting System (Go)

JSON REST API for creating polls and casting votes. The stack is Go + chi + Postgres, with JWT authentication, worker pool, metrics, and Docker support.

## Requirements

- Go (64-bit, recommended 1.22+)
- Docker + Docker Compose

## Project structure

- `cmd/server` – entrypoint
- `internal/config` – env/config loading
- `internal/domain` – domain models and services
- `internal/repository/postgres` – Postgres repositories
- `internal/http` – router, handlers, middleware
- `internal/worker` – vote aggregation worker pool
- `internal/metrics` – Prometheus counters
- `internal/db/migrations` – SQL migrations
- `docs` – Swagger docs

## Running Postgres

```bash
docker compose up -d db
```

## Migrations (golang-migrate CLI)

```bash
migrate -path internal/db/migrations \
  -database "postgres://polling_user:polling_pass@localhost:5432/polling_db?sslmode=disable" \
  up
```

Tables: `users`, `polls`, `options`, `votes`, `aggregated_results`, indexes, and a seeded admin (`admin@example.com`).

## Run the API locally

```bash
export APP_PORT=8080
export DB_DSN="postgres://polling_user:polling_pass@localhost:5432/polling_db?sslmode=disable"
export JWT_SECRET="super-secret-change-me"
go run ./cmd/server
```

Health check: `curl http://localhost:8080/health`

Swagger UI: `http://localhost:8080/swagger/index.html`

Prometheus metrics: `http://localhost:8080/metrics`

## Docker

Build and run the app container (expects the `db` compose service):

```bash
docker build -t polling-system .
docker run --rm -p 8080:8080 --env DB_DSN="postgres://polling_user:polling_pass@db:5432/polling_db?sslmode=disable" polling-system
```

## Testing

```bash
go test ./...
```

## API overview

Public:
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

Authenticated:
- `GET  /api/v1/polls`
- `GET  /api/v1/polls/{id}`
- `POST /api/v1/polls/{id}/vote`
- `GET  /api/v1/polls/{id}/results`

Admin-only:
- `POST  /api/v1/polls`
- `PATCH /api/v1/polls/{id}/status`
- `GET   /api/v1/users`
- `PATCH /api/v1/users/{id}/role`

## Behavior and reliability

- JWT auth with roles (`admin`, `user`), bcrypt password hashing.
- Voting is idempotent per poll/user via DB unique constraint; duplicate votes return HTTP 409.
- Poll must be `active` to accept votes; `draft`/`closed` votes are rejected.
- In-memory cache (10s TTL) for poll results with invalidation on new votes.
- Rate limiting on the vote endpoint (per-IP limiter) plus CORS and structured request logging.
- Worker pool consumes vote events and updates aggregated results with retry + backoff.
- Prometheus counter `polling_http_requests_total` (method/path/status) exposed at `/metrics`.
