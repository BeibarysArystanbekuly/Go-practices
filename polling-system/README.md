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

3. Apply migrations (run the SQL in `internal/db/migrations/*.sql`
   using any Postgres client connecting to:

   - host: `localhost`
   - port: `5432`
   - user: `polling_user`
   - password: `polling_pass`
   - database: `polling_db`

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

Admin-only:

- `POST  /api/v1/polls`
- `PATCH /api/v1/polls/{id}/status`
- `GET   /api/v1/users`
- `PATCH /api/v1/users/{id}/role`
