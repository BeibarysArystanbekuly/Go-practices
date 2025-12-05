# Polling / Voting System (Go)

Simple REST API for polls and voting, built in Go with Postgres and Docker.

## Requirements

- Go (64-bit, recommended 1.22+)
- Docker + Docker Compose

## Quick start

1. Start Postgres:

   ```bash
   docker compose up -d db
   ```

2. Install Go dependencies:

   ```bash
   go mod tidy
   ```

3. Apply migrations (recommended: [golang-migrate](https://github.com/golang-migrate/migrate)):

   ```bash
   migrate -path internal/db/migrations -database "postgres://polling_user:polling_pass@localhost:5432/polling_db?sslmode=disable" up
   ```

   The migrations create users/polls/options/votes plus aggregation tables and indexes; a default admin (`admin@example.com` / password hash) is seeded.

4. Run the server:

   ```bash
   go run ./cmd/server
   ```

5. Test health endpoint:

   ```bash
   curl http://localhost:8080/health
   ```

## Basic API

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET  /api/v1/polls`
- `GET  /api/v1/polls/{id}`
- `POST /api/v1/polls/{id}/vote`
- `GET  /api/v1/polls/{id}/results`
- `GET  /metrics` (Prometheus-style plain-text metrics)

Admin-only:

- `POST  /api/v1/polls`
- `PATCH /api/v1/polls/{id}/status`
- `GET   /api/v1/users`
- `PATCH /api/v1/users/{id}/role`

### Notes on behavior

- JWT auth (roles: `admin`, `user`), password hashing (bcrypt).
- Vote endpoint is rate-limited per user to reduce spam; duplicate votes are rejected idempotently.
- Background worker pool consumes vote events and maintains aggregated results with retry/backoff.
- Results endpoint caches responses briefly to offload the database.
