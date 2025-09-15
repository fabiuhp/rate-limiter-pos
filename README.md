# Go Rate Limiter (IP and Token)

Objective: A configurable rate limiter in Go that limits requests per second by IP or by access token, with Redis persistence and an easy-to-swap storage strategy. Runs as an HTTP middleware on port `8080`.

## Features

- Limit by IP and/or by Token (`API_KEY` header).
- Token rules override IP rules when present.
- Configurable requests-per-second and block duration.
- Returns HTTP 429 with message when exceeded.
- All limiter state persisted in Redis (via `STORE_DRIVER=redis`).
- Pluggable storage strategy (Redis, in-memory for tests).
- Middleware separated from core limiter logic.

## Quick Start (Docker Compose)

1. Adjust `.env` if needed.
2. Build and run:
   - `docker-compose up --build`
3. Test:
   - `curl -i http://localhost:8080/`
   - Provide a token: `curl -i -H 'API_KEY: abc123' http://localhost:8080/`

## Configuration

Environment variables (can be placed in `.env`):

- `SERVER_PORT`: HTTP port (default `8080`).
- `RATE_LIMIT_STRATEGY`: `ip`, `token`, or `both` (default `both`).
- `API_KEY_HEADER`: Header name for token (default `API_KEY`).
- `RATE_LIMIT_PER_SECOND`: Default RPS per IP (default `5`).
- `RATE_LIMIT_BLOCK_DURATION`: Block duration for IP after exceeding (default `5m`).
- `TOKEN_DEFAULT_LIMIT_PER_SECOND`: Default token RPS when `strategy=token` and token is not listed.
- `TOKEN_DEFAULT_BLOCK_DURATION`: Default token block duration.
- `TOKEN_LIMITS`: Per-token rules `token:limit:blockDuration` (e.g. `abc123:100:5m,xyz:50:1m`).
- `STORE_DRIVER`: `redis` (default) or `memory`.
- `REDIS_ADDR`: Redis address (default `redis:6379` in Compose).
- `REDIS_DB`, `REDIS_PASSWORD`: Optional.

Notes:
- If `strategy=both` (default): when a valid token has a configured rule, it overrides IP limits; otherwise IP limits apply.
- If `strategy=ip`: only IP limits apply (token ignored).
- If `strategy=token`: token limits apply; if a token rule isn't found, the optional token default is used if >0.

## How It Works

- Middleware (`internal/middleware`) calls the limiter (`internal/limiter`) for each request.
- Limiter chooses the rule (token overrides IP) and calls the storage with `(scope, key, limit, window=1s, block)`.
- Redis store uses a Lua script to increment a 1-second window counter and set a block key when exceeded.
- When blocked, responses return HTTP 429 and message: `you have reached the maximum number of requests or actions allowed within a certain time frame`.

## Development

- Run locally:
  - `go run ./cmd/server` (requires Redis `REDIS_ADDR` reachable)
- Tests:
  - `go test ./...` (uses in-memory store)

## Project Layout

- `cmd/server`: HTTP server wiring.
- `internal/config`: env and .env loader; rules parsing.
- `internal/limiter`: rule selection; request evaluation.
- `internal/middleware`: HTTP middleware.
- `internal/storage`: Store interface; Redis and in-memory implementations.
