# Nust Devs Platform — Milestones

> Track NUST developer activity, contributions, and top projects using the GitHub public API.
>
> **Stack:** Go API + Next.js dashboard · PostgreSQL · Redis · Asynq workers
>
> **Discovery model:** Hybrid — manual admin registry first, then `@nust.edu.pk` email + GitHub org verification.

---

## Progress Overview

| Milestone | Name | Target | Status |
|-----------|------|--------|--------|
| M0 | Foundation | Week 1 | ⬜ Not started |
| M1 | Database & Developer Registry | Week 2 | ⬜ Not started |
| M2 | GitHub Client & Rate Limiter | Week 2–3 | ⬜ Not started |
| M3 | Sync Worker | Week 3–4 | ⬜ Not started |
| M4 | Stats Engine & Public API | Week 4–5 | ⬜ Not started |
| M5 | Public Dashboard | Week 5–7 | ⬜ Not started |
| M6 | Hybrid Verification | Week 7–8 | ⬜ Not started |
| M7 | Production Hardening | Week 8–9 | ⬜ Not started |
| M8 | Future Enhancements | Post-v1 | ⬜ Backlog |

**Legend:** ⬜ Not started · 🔄 In progress · ✅ Done

---

## Milestone 0 — Foundation

**Target:** Week 1  
**Goal:** Runnable monorepo skeleton with CI and local dev environment.

### Tasks

- [ ] Initialize Go module and project layout (`cmd/`, `internal/`)
- [ ] Docker Compose: PostgreSQL, Redis, API, worker, web
- [ ] Config loading (env vars + `.env.example`)
- [ ] Structured logging (slog or zerolog)
- [ ] Health endpoint (`GET /health`)
- [ ] SQL migrations tooling (golang-migrate)
- [ ] GitHub Actions: lint, test, build
- [ ] `.gitignore` and expanded README with setup instructions

### Exit Criteria

- [ ] `docker compose up` starts all services without errors
- [ ] `GET /health` returns `200 OK`
- [ ] CI pipeline passes on push

### Key Files

```
cmd/server/main.go
cmd/worker/main.go
docker-compose.yml
.env.example
.github/workflows/ci.yml
```

---

## Milestone 1 — Database & Developer Registry

**Target:** Week 2  
**Goal:** Persist developers; admin can register them.

### Tasks

- [ ] Design and apply initial migration
  - [ ] `developers`
  - [ ] `repos`
  - [ ] `developer_repos`
  - [ ] `developer_snapshots`
  - [ ] `contribution_days`
  - [ ] `sync_jobs`
  - [ ] `admin_users`
- [ ] Repository layer (pgx or sqlc)
- [ ] Admin auth (JWT + bcrypt, seed admin user)
- [ ] CRUD endpoints: register / list / update / delete developers
- [ ] Basic Next.js shell with admin login + add-developer form

### Exit Criteria

- [ ] Admin can log in with seeded credentials
- [ ] Admin can add a GitHub username
- [ ] Developer persists in DB and appears in admin list
- [ ] Admin can update and delete developers

### API Endpoints

| Method | Path | Auth |
|--------|------|------|
| POST | `/api/v1/admin/auth/login` | Public |
| POST | `/api/v1/admin/developers` | Admin |
| GET | `/api/v1/admin/developers` | Admin |
| PATCH | `/api/v1/admin/developers/:id` | Admin |
| DELETE | `/api/v1/admin/developers/:id` | Admin |

---

## Milestone 2 — GitHub Client & Rate Limiter

**Target:** Week 2–3  
**Goal:** Rock-solid GitHub integration layer.

### Tasks

- [ ] REST client
  - [ ] `GET /users/{username}` — profile
  - [ ] `GET /users/{username}/repos` — paginated repos
  - [ ] `GET /repos/{owner}/{repo}/languages` — language breakdown
- [ ] GraphQL client
  - [ ] `contributionsCollection` — contribution calendar
- [ ] Central rate-limit manager
  - [ ] Track `x-ratelimit-remaining` and `x-ratelimit-reset`
  - [ ] Exponential backoff on 403/429
  - [ ] Pause when remaining quota < 100
- [ ] Unit tests with mocked GitHub responses
- [ ] Integration test against live GitHub API (skipped in CI without token)

### Exit Criteria

- [ ] Test or CLI fetches a known user's profile and repos
- [ ] Rate limiter prevents unsafe burst requests
- [ ] All GitHub calls require `GITHUB_TOKEN` (no unauthenticated production use)

### Rate-Limit Rules (Non-Negotiable)

1. Always use a GitHub PAT (`GITHUB_TOKEN`) — 5,000 req/hr vs 60 unauthenticated
2. Read rate-limit headers on every response
3. Staggered sync tiers: active devs every 6h, inactive every 24h
4. Persist everything — never re-fetch unless stale
5. Exponential backoff on 403/429
6. GraphQL batching — one contribution calendar query per user

---

## Milestone 3 — Sync Worker

**Target:** Week 3–4  
**Goal:** Background jobs populate DB from GitHub.

### Tasks

- [ ] Asynq job queue setup (enqueue, process, retry, dead letter)
- [ ] Sync job pipeline: profile → repos → languages → contribution days
- [ ] Daily snapshot writer (for trend graphs)
- [ ] Staggered scheduling
  - [ ] Full sync nightly
  - [ ] Incremental sync every 6h
- [ ] Admin trigger sync endpoint (single developer or all)
- [ ] Admin sync status endpoint (queue depth + rate limit state)
- [ ] Idempotent upserts; track `last_synced_at` per developer

### Exit Criteria

- [ ] Register 5 developers → worker syncs all automatically
- [ ] DB contains profiles, repos, languages, and contribution data
- [ ] Daily snapshots are written for trend tracking
- [ ] Manual sync trigger works from admin API

### API Endpoints

| Method | Path | Auth |
|--------|------|------|
| POST | `/api/v1/admin/sync` | Admin |
| GET | `/api/v1/admin/sync/status` | Admin |

---

## Milestone 4 — Stats Engine & Public API

**Target:** Week 4–5  
**Goal:** Computed metrics and public read endpoints.

### Tasks

- [ ] Activity score computation service
- [ ] Leaderboard queries (stars, commits, activity score, repos)
- [ ] Top projects aggregation (NUST dev repos ranked)
- [ ] Platform overview stats (totals, language breakdown)
- [ ] Public API handlers with pagination, sorting, and Redis caching
- [ ] OpenAPI spec (swaggo or manual)

### Activity Score Formula (v1)

```
activity_score =
  (commits_last_90d * 3) +
  (public_repos * 2) +
  (total_stars * 0.1) +
  (recent_pushes_last_30d * 5)
```

Weights configurable in admin settings. Recompute after each sync.

### Exit Criteria

- [ ] All public API endpoints return real data from synced developers
- [ ] Leaderboards sort correctly by each metric
- [ ] Redis cache reduces repeated DB load for hot endpoints

### API Endpoints

| Method | Path | Auth |
|--------|------|------|
| GET | `/api/v1/developers` | Public |
| GET | `/api/v1/developers/:username` | Public |
| GET | `/api/v1/developers/:username/repos` | Public |
| GET | `/api/v1/developers/:username/contributions` | Public |
| GET | `/api/v1/leaderboard` | Public |
| GET | `/api/v1/projects/top` | Public |
| GET | `/api/v1/stats/overview` | Public |
| GET | `/api/v1/stats/languages` | Public |

---

## Milestone 5 — Public Dashboard

**Target:** Week 5–7  
**Goal:** Beautiful, data-rich frontend.

### Tasks

- [ ] Design system (Tailwind + shadcn/ui, dark/light mode)
- [ ] Home page — overview stats, top devs, top projects, language chart
- [ ] Developer list — searchable/filterable grid
- [ ] Developer profile — heatmap, language chart, repo table, stats cards
- [ ] Leaderboard — sortable table with sparklines
- [ ] Projects page — top repos grid/table
- [ ] Stats page — platform-wide charts and trends
- [ ] Responsive layout, loading states, error boundaries
- [ ] SEO metadata (Open Graph for developer profiles)

### Chart Types

- [ ] Contribution heatmap (GitHub-style calendar)
- [ ] Language distribution (donut/bar)
- [ ] Activity over time (line/area from daily snapshots)
- [ ] Leaderboard sparklines
- [ ] Top projects bar chart
- [ ] Stars/forks growth (from repo snapshots)

### Exit Criteria

- [ ] All pages render live data from the API
- [ ] Charts update after a sync completes
- [ ] Mobile layout is usable
- [ ] Profile pages have correct Open Graph tags

### Pages

| Route | Description |
|-------|-------------|
| `/` | Platform overview |
| `/developers` | Developer directory |
| `/developers/[username]` | Developer profile |
| `/leaderboard` | Rankings |
| `/projects` | Top repositories |
| `/stats` | Platform-wide analytics |
| `/admin` | Admin panel |

---

## Milestone 6 — Hybrid Verification

**Target:** Week 7–8  
**Goal:** Expand beyond manual registry.

### Tasks

- [ ] Email verification — GitHub public email matches `@nust.edu.pk`
- [ ] Org verification — sync members from configured NUST GitHub org
- [ ] Verification badges on frontend
- [ ] Filter: "verified NUST developers only" on leaderboard and lists
- [ ] Admin UI: run verification checks, override status

### Verification Statuses

| Status | Meaning |
|--------|---------|
| `registered` | Admin added, unverified |
| `email_verified` | GitHub email matches `@nust.edu.pk` |
| `org_verified` | Member of configured NUST GitHub org |
| `manual_verified` | Admin manually confirmed |

### Exit Criteria

- [ ] Developers auto-verified when email or org matches
- [ ] Verification badges visible on profiles and leaderboard
- [ ] "Verified only" filter works on public pages
- [ ] Admin can override verification status

---

## Milestone 7 — Production Hardening

**Target:** Week 8–9  
**Goal:** Deployable, monitored, documented.

### Tasks

- [ ] Production Dockerfile (multi-stage Go build + Next.js standalone)
- [ ] Environment-specific config (dev / staging / prod)
- [ ] Rate limit monitoring in admin dashboard
- [ ] Error alerting hooks (log aggregation ready)
- [ ] Database indexes for leaderboard and snapshot queries
- [ ] API response caching strategy (Redis TTLs per endpoint)
- [ ] Comprehensive README: architecture, API docs, deployment guide
- [ ] Seed script for demo data

### Exit Criteria

- [ ] Platform deployable via Docker in a single command
- [ ] Another developer can run locally using README alone
- [ ] Leaderboard and stats queries perform well at 50+ developers
- [ ] Admin can monitor GitHub API quota usage

---

## Milestone 8 — Future Enhancements (Post-v1 Backlog)

**Goal:** Track ideas without blocking v1 launch.

- [ ] PR/issue stats via GraphQL (per-repo)
- [ ] Cohort/year filters (batch of NUST graduates)
- [ ] Developer self-registration + claim flow
- [ ] GitHub OAuth login for developers to manage their profile
- [ ] Export stats (CSV/JSON)
- [ ] Webhooks for real-time repo events (if org-owned)
- [ ] Compare two developers side-by-side
- [ ] Monthly "NUST Dev of the Month" automated ranking
- [ ] Public API rate limiting for third-party consumers

---

## v1 Success Metrics

| Metric | Target |
|--------|--------|
| Registered & synced developers | 50+ |
| Public page load time | < 2s |
| Sync reliability | No 429 errors within rate limits |
| Live chart types | 6+ |
| Admin workflow | Add / sync / verify end-to-end |

---

## Risk Register

| Risk | Impact | Mitigation |
|------|--------|------------|
| GitHub rate limits with many developers | Sync stalls | Staggered sync, Redis queue, PAT, aggressive persistence |
| Events API 90-day limit | Incomplete history | GraphQL contribution calendar + daily snapshots |
| Private repos invisible | Incomplete stats | Document clearly; only public data |
| Fake NUST affiliation | Bad data | Hybrid verification (email + org + admin manual) |
| GitHub API changes | Broken sync | Version-pin API calls; abstract client interface |
| Large repo lists per user | Slow sync | Paginate; cap repo detail fetch to top N by stars/activity |

---

## Implementation Order

```
M0 Foundation
  └─► M1 Registry
        └─► M2 GitHub Client
              └─► M3 Sync Worker
                    └─► M4 Stats API
                          └─► M5 Dashboard  ← usable public platform
                                └─► M6 Verification
                                      └─► M7 Production  ← v1 launch
                                            └─► M8 Backlog
```

Work strictly milestone-by-milestone. Do not start the next milestone until exit criteria for the current one are met.

---

## Repository Layout (Target)

```
Nust-devs/
├── cmd/
│   ├── server/main.go
│   └── worker/main.go
├── internal/
│   ├── config/
│   ├── github/
│   ├── models/
│   ├── repository/
│   ├── service/
│   ├── handler/
│   └── worker/
├── migrations/
├── web/
│   ├── app/
│   ├── components/
│   └── lib/api.ts
├── docker-compose.yml
├── milestones.md
├── go.mod
└── README.md
```

---

*Last updated: May 2026*
