# Project: Global-Ranks

## What This Project Does

Multi-game leaderboard platform. Go backend + PostgreSQL. Clients identify via UUID,
submit scores, fetch leaderboards, edit usernames. Server generates identicon avatars.

## Stack

- **Language:** Go 1.22+ (uses net/http routing patterns)
- **Database:** PostgreSQL
- **Build:** `make build`
- **Key deps:** pgx/v5, google/uuid, x/time/rate (3 external only)

## Directory Layout

```text
cmd/global-ranks/main.go     → Entry point, wiring, graceful shutdown
internal/config/              → Env-var configuration
internal/database/            → DB connection pool, migration runner
internal/handler/             → HTTP handlers (score, leaderboard, user, avatar)
internal/middleware/           → Logging, recovery, rate limiting
internal/model/               → Data structs + DB queries
internal/identicon/           → SHA256-based identicon PNG generator
internal/validate/            → Input validation (scores, usernames)
migrations/                   → Embedded SQL schema migrations
```

## Essential Commands

```bash
# Build
make build

# Run
GR_DB_DSN="postgres://user:pass@localhost/globalranks" make run

# Test
make test

# Lint (must pass before committing)
make lint   # runs: go vet && staticcheck
```

## API Endpoints (base: /api/v1)

- `GET  /health` — health + DB ping
- `POST /register` — register new user, returns secret token (once)
- `POST /sessions` — start game session (auth required)
- `POST /scores` — submit score for a session (auth required)
- `GET  /games/{slug}/leaderboard` — paginated best-per-user
- `GET  /users/{uuid}` — user profile (public)
- `PATCH /users/{uuid}` — update username (auth required, own UUID only)
- `GET  /avatars/{uuid}.png` — deterministic identicon

Auth: `X-User-UUID` + `X-Auth-Token` headers on protected endpoints.

## Configuration (env vars)

`GR_DB_DSN`, `GR_PORT` (8080), `GR_RATE_SCORE_PER_SEC` (0.2),
`GR_RATE_SCORE_BURST` (3), `GR_RATE_GENERAL_PER_SEC` (10),
`GR_RATE_GENERAL_BURST` (30), `GR_MAX_SCORE` (999999999)

## Project-Specific Rules

- Server is publicly available, expect many client connections.
- Rate limiting is dual-layer: per-IP (general) + per-UUID (score submissions).
- Games are auto-created on first session creation for that slug.
- Users register via POST /register, which returns a secret token (once).
- Score submission requires a single-use game session for anti-manipulation.
- Token auth (X-User-UUID + X-Auth-Token) protects score, session, and username endpoints.

## Skills Available

- `codebase-navigator` — use when first exploring this repo
- `code-quality` — use before committing any changes

## See Also

@README.md
