# Polling / Voting System (Go)

JSON REST API for creating polls and casting votes. The stack is Go + chi + Postgres, with JWT authentication, worker pool, metrics, and Docker support.

## Requirements

- Go (64-bit, recommended 1.22+)
- Docker + Docker Compose

## Project structure

- `cmd/server` - entrypoint
- `internal/config` - env/config loading
- `internal/domain` – domain models and services
- `internal/repository/postgres` – Postgres repositories
- `internal/http` – router, handlers, middleware
- `internal/worker` – vote aggregation worker pool
- `internal/metrics` – Prometheus counters
- `internal/db/migrations` – SQL migrations
- `docs` - Swagger docs

## Quickstart (Docker Compose)

```bash
docker compose up -d
```

Services:
- `api` – builds the Go server and exposes port `8080`
- `db` – Postgres with default creds (`polling_user` / `polling_pass`)

Environment (override as needed):
- `APP_PORT` (default `8080`)
- `DB_DSN` (defaults to the compose DB: `postgres://polling_user:polling_pass@db:5432/polling_db?sslmode=disable`)
- `JWT_SECRET` (set your own)
- `JWT_ISSUER` (default `polling-system`)
- Seeded admin: `admin@example.com` (password hash in `2_seed_admin.up.sql` — change if needed)

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
export JWT_ISSUER="polling-system"
go run ./cmd/server
```

Health check: `curl http://localhost:8080/health`
Readiness: `curl http://localhost:8080/ready`

Swagger UI: `http://localhost:8080/swagger/index.html`

Prometheus metrics: `http://localhost:8080/metrics`

## Docker (manual build/run)

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
- `GET  /health`
- `GET  /ready`

Admin-only:
- `POST  /api/v1/polls`
- `PATCH /api/v1/polls/{id}`
- `PATCH /api/v1/polls/{id}/status`
- `DELETE /api/v1/polls/{id}`
- `GET   /api/v1/users`
- `PATCH /api/v1/users/{id}/role`
- `PATCH /api/v1/users/{id}/deactivate`

## Error format

All errors are JSON:

```json
{"error": "invalid_input", "message": "title is required"}
```

Status mapping:
- `400` – validation / bad input
- `401` – invalid token / credentials / inactive user
- `403` – RBAC failures
- `404` – entity not found
- `409` – conflicts (e.g., duplicate vote)
- `500/503` – unexpected / dependency unavailable

## Behavior and reliability

- JWT auth with roles (`admin`, `user`), bcrypt password hashing.
- Inactive users are rejected at login (`is_active=false`).
- Voting is idempotent per poll/user via DB unique constraint; duplicate votes return HTTP 409.
- Poll must be `active` to accept votes; `draft`/`closed` votes are rejected.
- Options are validated against the poll by composite FK and service errors.
- In-memory cache (10s TTL) for poll results with invalidation on new votes.
- Rate limiting on the vote endpoint (per-IP limiter) plus CORS and structured request logging.
- Worker pool consumes vote events and updates aggregated results with retry + backoff.
- Prometheus counter `polling_http_requests_total` (method/path/status) exposed at `/metrics`.
- Graceful shutdown handles SIGINT/SIGTERM and drains the worker pool.

## Example curl calls

```bash
# Login (admin)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"<your-admin-password>"}'

# Create a poll (admin token)
curl -X POST http://localhost:8080/api/v1/polls \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Lunch?","options":["Pizza","Salad"]}'

# Vote (user token)
curl -X POST http://localhost:8080/api/v1/polls/1/vote \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"option_id":2}'
```
