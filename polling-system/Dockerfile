FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download || true

COPY . .
RUN go build -o server ./cmd/server

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/server .
ENV APP_PORT=8080
ENV DB_DSN=postgres://polling_user:polling_pass@db:5432/polling_db?sslmode=disable
ENV JWT_SECRET=super-secret-change-me

EXPOSE 8080

CMD ["./server"]
