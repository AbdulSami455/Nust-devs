# NUST Devs

Track GitHub activity, contributions, and top projects from NUST developers.

**Stack:** Go API · Next.js dashboard · PostgreSQL · Redis · Asynq workers

---

## Quick Start (Docker)

```bash
cp .env.example .env          # add your GITHUB_TOKEN
docker compose up --build
```

- Frontend: http://localhost:3000
- API:      http://localhost:8080
- Admin:    http://localhost:3000/admin  (admin@nust.edu.pk / admin123)

---

## Local Development

### Prerequisites

- Go 1.24+
- Node.js 22+
- Docker (for Postgres + Redis)

### 1. Start dependencies

```bash
docker run -d --name nust-postgres \
  -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=nustdevs \
  -p 5432:5432 postgres:16-alpine

docker run -d --name nust-redis -p 6379:6379 redis:7-alpine
```

### 2. Configure environment

```bash
cp .env.example .env
# Set GITHUB_TOKEN to a GitHub PAT with read:user scope
```

### 3. Run the API (runs migrations on start)

```bash
go run ./cmd/server
```

### 4. Run the worker (separate terminal)

```bash
go run ./cmd/worker
```

### 5. Run the frontend

```bash
cd web && npm install && npm run dev
```

---

## Architecture

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│  Next.js    │───▶│  Go API      │───▶│  PostgreSQL │
│  (port 3000)│    │  (port 8080) │    │  (port 5432)│
└─────────────┘    └──────┬───────┘    └─────────────┘
                          │
                   ┌──────▼───────┐    ┌─────────────┐
                   │  Asynq       │───▶│  Redis      │
                   │  Worker      │    │  (port 6379)│
                   └──────┬───────┘    └─────────────┘
                          │
                   ┌──────▼───────┐
                   │  GitHub API  │
                   │  (REST+GQL)  │
                   └──────────────┘
```

### Key packages

| Path | Purpose |
|------|---------|
| `cmd/server/` | HTTP API entrypoint |
| `cmd/worker/` | Asynq background worker + scheduler |
| `internal/github/` | GitHub REST + GraphQL client with rate limiting |
| `internal/service/sync.go` | Full sync pipeline: profile → repos → contributions → score |
| `internal/repository/` | pgx database queries |
| `internal/cache/` | Redis JSON cache |
| `internal/handler/` | HTTP handlers (public + admin) |
| `internal/worker/` | Asynq task definitions and processors |
| `migrations/` | golang-migrate SQL migration files |
| `web/` | Next.js 15 App Router frontend |

---

## API Reference

### Public endpoints (no auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/developers` | List developers (paginated, sorted by activity score) |
| GET | `/api/v1/developers/:username` | Developer profile |
| GET | `/api/v1/developers/:username/repos` | Developer repositories |
| GET | `/api/v1/developers/:username/contributions` | Contribution calendar (last 365 days) |
| GET | `/api/v1/developers/:username/rank-insight` | AI rank and badge explanation |
| GET | `/api/v1/leaderboard?sort_by=activity_score` | Rankings |
| GET | `/api/v1/projects/top` | Top repositories by stars |
| GET | `/api/v1/repos/:id/summary` | AI project impact summary |
| GET | `/api/v1/stats/overview` | Platform totals |
| GET | `/api/v1/stats/languages` | Language breakdown |

### Admin endpoints (Bearer token required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/admin/auth/login` | Login → JWT |
| GET | `/api/v1/admin/developers` | List all developers |
| POST | `/api/v1/admin/developers` | Register developer |
| PATCH | `/api/v1/admin/developers/:id` | Update developer |
| DELETE | `/api/v1/admin/developers/:id` | Remove developer |
| POST | `/api/v1/admin/sync` | Trigger sync (all or `?id=uuid`) |
| GET | `/api/v1/admin/sync/status` | Queue depth + GitHub rate limit |

### Activity Score Formula

```
score = (commits_last_90d × 3) + (public_repos × 2) + (total_stars × 0.1) + (commits_last_30d × 5)
```

Recomputed after every sync.

---

## GitHub Rate Limiting

- Requires a GitHub PAT — 5,000 req/hr authenticated vs 60 unauthenticated
- Headers tracked on every response (`x-ratelimit-remaining`, `x-ratelimit-reset`)
- Pauses automatically when quota < 100
- Exponential backoff on 403/429
- Worker concurrency capped at 3
- Schedule: all developers nightly, active devs every 6h

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | API listen port |
| `DATABASE_URL` | — | PostgreSQL DSN |
| `REDIS_URL` | `redis://localhost:6379` | Redis address |
| `GITHUB_TOKEN` | — | **Required.** GitHub PAT |
| `JWT_SECRET` | — | **Required for API server.** Admin JWT signing secret, minimum 32 characters |
| `ADMIN_EMAIL` | — | Required only when seeding the first admin user |
| `ADMIN_PASSWORD` | — | Required with `ADMIN_EMAIL`, minimum 12 characters |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,http://127.0.0.1:3000` | Comma-separated frontend origins allowed to call the API with credentials. Wildcards are rejected |
| `SECURE_COOKIES` | `true` | Set to `false` only for local HTTP development |
| `PUBLIC_RATE_LIMIT_REQUESTS` | `600` | Max public API requests per client per window |
| `PUBLIC_RATE_LIMIT_WINDOW` | `1m` | Public API rate limit window as a Go duration |
