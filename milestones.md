# Nust Devs Platform — Milestones

> Track NUST developer activity, contributions, and top projects using the GitHub public API.
>
> **Stack:** Go API + Next.js dashboard · PostgreSQL · Redis · Asynq workers
>
> **Discovery model:** Admin can register developers directly (**GitHub username** required; email optional). NUST developers can also **submit a join request** at `/join` for admin approval. Email domain verification (M6) and full self-registration with OAuth (M8) remain deferred.

---

## Progress Overview

| Milestone | Name | Target | Status |
|-----------|------|--------|--------|
| M0 | Foundation | Week 1 | ✅ Done |
| M1 | Database & Developer Registry | Week 2 | ✅ Done |
| M2 | GitHub Client & Rate Limiter | Week 2–3 | ✅ Done |
| M3 | Sync Worker | Week 3–4 | ✅ Done |
| M4 | Stats Engine & Public API | Week 4–5 | ✅ Done |
| M5 | Public Dashboard | Week 5–7 | ✅ Done |
| M5b | Frontend Revamp & Innovation Graph | Post-M5 | ✅ Done |
| M7 | Production Hardening | Week 7–8 | 🔄 In progress |
| M6 | Email Verification (optional) | Week 8+ | ⬜ Low priority |
| M8 | Future Enhancements | Post-v1 | 🔄 Partial (claim flow done) |

**Legend:** ⬜ Not started · 🔄 In progress · ✅ Done

---

## Milestone 0 — Foundation

**Target:** Week 1  
**Goal:** Runnable monorepo skeleton with CI and local dev environment.

### Tasks

- [x] Initialize Go module and project layout (`cmd/`, `internal/`)
- [x] Config loading (env vars + `.env.example`)
- [x] Structured logging (slog or zerolog)
- [x] Health endpoint (`GET /health`)
- [x] SQL migrations tooling (golang-migrate)
- [ ] GitHub Actions: lint, test, build
- [x] `.gitignore` and expanded README with setup instructions

### Exit Criteria

- [x] `GET /health` returns `200 OK`
- [ ] CI pipeline passes on push

### Key Files

```
cmd/server/main.go
cmd/worker/main.go
.env.example
```

---

## Milestone 1 — Database & Developer Registry

**Target:** Week 2  
**Goal:** Persist developers; admin can register them with GitHub username (email optional).

### Tasks

- [x] Design and apply initial migration
  - [x] `developers` (includes `github_username`, optional `email`, optional `display_name`, `notes`)
  - [x] `repos`
  - [x] `developer_repos`
  - [x] `developer_snapshots`
  - [x] `contribution_days`
  - [x] `sync_jobs`
  - [x] `admin_users`
  - [x] `developer_requests` (join requests, pending/approved/rejected)
  - [x] `repos.license` column (SPDX from GitHub sync)
- [x] Repository layer (pgx)
- [x] Admin auth (JWT + bcrypt, seed admin user)
- [x] CRUD endpoints: register / list / update / delete developers
- [x] Duplicate GitHub username prevention (normalized lowercase, case-insensitive check)
- [x] Basic Next.js shell with admin login + add-developer form
- [x] Public join request flow (`/join`) with admin approve/reject

### Exit Criteria

- [x] Admin can log in with seeded credentials
- [x] Admin can add a developer with **GitHub username** (email optional)
- [x] Developer persists in DB and appears in admin list
- [x] Admin can update and delete developers
- [x] Duplicate usernames are rejected (admin add + join request)
- [x] Pending join requests appear in admin dashboard; approve creates developer + enqueues sync

### Register Developer Payload (v1)

| Field | Required | Notes |
|-------|----------|-------|
| `github_username` | Yes | Normalized lowercase; used for GitHub sync |
| `email` | No | Stored if provided; domain verification deferred to M6 |
| `display_name` | No | Optional override |
| `notes` | No | Admin-only |

### API Endpoints

| Method | Path | Auth |
|--------|------|------|
| POST | `/api/v1/admin/auth/login` | Public |
| POST | `/api/v1/admin/developers` | Admin |
| GET | `/api/v1/admin/developers` | Admin |
| PATCH | `/api/v1/admin/developers/:id` | Admin |
| DELETE | `/api/v1/admin/developers/:id` | Admin |
| POST | `/api/v1/profile-requests` | Public |
| GET | `/api/v1/profile-requests/check` | Public |
| GET | `/api/v1/admin/profile-requests` | Admin |
| POST | `/api/v1/admin/profile-requests/:id/approve` | Admin |
| POST | `/api/v1/admin/profile-requests/:id/reject` | Admin |

---

## Milestone 2 — GitHub Client & Rate Limiter

**Target:** Week 2–3  
**Goal:** Rock-solid GitHub integration layer.

### Tasks

- [x] REST client
  - [x] `GET /users/{username}` — profile
  - [x] `GET /users/{username}/repos` — paginated repos
  - [x] `GET /repos/{owner}/{repo}/languages` — language breakdown
  - [x] Repo `license` (SPDX) on sync
- [x] GraphQL client
  - [x] `contributionsCollection` — contribution calendar
- [x] Central rate-limit manager
  - [x] Track `x-ratelimit-remaining` and `x-ratelimit-reset`
  - [x] Exponential backoff on 403/429
  - [x] Pause when remaining quota < 100
- [x] Unit tests with mocked GitHub responses
- [ ] Integration test against live GitHub API (skipped in CI without token)

### Exit Criteria

- [x] Test or CLI fetches a known user's profile and repos
- [x] Rate limiter prevents unsafe burst requests
- [x] All GitHub calls require `GITHUB_TOKEN` (no unauthenticated production use)

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

- [x] Asynq job queue setup (enqueue, process, retry, dead letter)
- [x] Sync job pipeline: profile → repos → languages → contribution days
- [x] Daily snapshot writer (for trend graphs)
- [x] Staggered scheduling
  - [x] Full sync nightly
  - [x] Incremental sync every 6h
- [x] Admin trigger sync endpoint (single developer or all)
- [x] Admin sync status endpoint (rate limit state)
- [x] Idempotent upserts; track `last_synced_at` per developer
- [x] Redis cache invalidation after successful sync
- [x] Auto-sync enqueue on approved join request

### Exit Criteria

- [x] Register 5+ developers → worker syncs all automatically
- [x] DB contains profiles, repos, languages, and contribution data
- [x] Daily snapshots are written for trend tracking
- [x] Manual sync trigger works from admin API

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

- [x] Activity score computation service
- [x] Leaderboard queries (activity score, stars, repos, followers)
- [x] Top projects aggregation with filters (category, language, sort)
- [x] Platform overview stats (totals, language breakdown)
- [x] Community activity time series (`contribution_days` aggregate)
- [x] Developer spotlight endpoint
- [x] Live activity feed (recent repo pushes)
- [x] Open-source stats (original vs fork projects, contributors)
- [x] Innovation graph API (quarterly/monthly trends: pushes, repos, devs, orgs, languages, licenses)
- [x] Public API handlers with pagination, sorting, and Redis caching
- [ ] OpenAPI spec (swaggo or manual)

### Activity Score Formula (v1)

```
activity_score =
  (commits_last_90d * 3) +
  (public_repos * 2) +
  (total_stars * 0.1) +
  (commits_last_30d * 5)
```

Recomputed after each sync. (Note: formula uses contribution days, not literal commits.)

### Exit Criteria

- [x] All public API endpoints return real data from synced developers
- [x] Leaderboards sort correctly by each metric
- [x] Redis cache reduces repeated DB load for hot endpoints

### API Endpoints

| Method | Path | Auth |
|--------|------|------|
| GET | `/api/v1/developers` | Public |
| GET | `/api/v1/developers/spotlight` | Public |
| GET | `/api/v1/developers/:username` | Public |
| GET | `/api/v1/developers/:username/repos` | Public |
| GET | `/api/v1/developers/:username/contributions` | Public |
| GET | `/api/v1/leaderboard` | Public |
| GET | `/api/v1/projects/top` | Public |
| GET | `/api/v1/activity/recent` | Public |
| GET | `/api/v1/stats/overview` | Public |
| GET | `/api/v1/stats/languages` | Public |
| GET | `/api/v1/stats/community-activity` | Public |
| GET | `/api/v1/stats/open-source` | Public |
| GET | `/api/v1/stats/innovation-graph` | Public |

---

## Milestone 5 — Public Dashboard

**Target:** Week 5–7  
**Goal:** Beautiful, data-rich frontend.

### Tasks

- [x] Design system (Tailwind + shadcn/ui, dark/light mode, NUST palette)
- [x] Global Cmd+K command palette
- [x] Home page — bento stats, community pulse chart, spotlight, top devs/projects, OSS panel, live activity
- [x] Developer list — searchable grid with mini dev cards
- [x] Developer profile — dev card, share (copy link / X), contribution heatmap, top repos
- [x] Leaderboard — podium for top 3, tabs (Top 10/50/All), sort pills
- [x] Projects page — category filters (All / Original OSS / Forks), language & sort filters
- [x] Stats page — platform-wide charts
- [x] Innovation graph page (`/innovation`) — LebHub-style quarterly trends + live activity tab
- [x] Join page (`/join`) — submit profile request for admin approval
- [x] Mobile bottom navigation
- [x] Skeleton loading states (no raw "Loading…" on main dashboards)
- [x] SEO metadata (Open Graph for developer profiles)
- [ ] Error boundaries (global)

### Chart Types

- [x] Contribution heatmap (GitHub-style calendar on profile)
- [x] Language distribution (donut/bar on stats + innovation graph)
- [x] Activity over time (area chart — community pulse + innovation graph pushes)
- [ ] Leaderboard sparklines
- [x] Top projects bar chart (stats page)
- [ ] Stars/forks growth (from repo snapshots over time)

### Exit Criteria

- [x] All pages render live data from the API
- [x] Charts update after a sync completes (cache invalidation + refresh)
- [x] Mobile layout is usable
- [x] Profile pages have correct Open Graph tags

### Pages

| Route | Description |
|-------|-------------|
| `/` | Platform overview (bento dashboard) |
| `/developers` | Developer directory |
| `/developers/[username]` | Developer profile + dev card |
| `/leaderboard` | Rankings with podium |
| `/projects` | Open-source projects with filters |
| `/stats` | Platform-wide analytics |
| `/innovation` | Innovation graph & trends |
| `/join` | Request to join (admin approval) |
| `/admin` | Admin login |
| `/admin/dashboard` | Developer + request management |

---

## Milestone 5b — Frontend Revamp (LebHub-inspired)

**Goal:** Modern community dashboard UX (see `web/REVAMP_PLAN.md`).

| Revamp milestone | Status |
|------------------|--------|
| M1 — Theme, Cmd+K, shell | ✅ Done |
| M2 — Home bento dashboard, community pulse, spotlight | ✅ Done |
| M3 — Dev profile card, heatmap, share actions | ✅ Done |
| M4 — Leaderboard podium + tabs | ✅ Done |
| M5 — Live feed, mobile nav, polish | ✅ Done |

---

## Milestone 6 — Email Verification (Optional, Low Priority)

**Target:** Week 8+ *(after v1 launch — not a blocker for M7)*  
**Goal:** Validate admin-provided emails against NUST-affiliated domains. **No GitHub org verification.**

> Build the core platform first (M0–M5, M7). Add email verification only when the dashboard and sync are stable.

### Tasks

- [ ] Configurable email domain allowlist (admin-managed)
  - [ ] Default domains: `nust.edu.pk`, `seecs.edu.pk`
  - [ ] Support adding/removing domains without code changes (e.g. `@nbs.edu.pk`, `@scme.nust.edu.pk`)
- [ ] Verify **admin-provided email** against allowlist (not GitHub org membership)
- [ ] Optional: cross-check with GitHub public email during sync if available
- [ ] Store matched domain on developer record (e.g. `verified_email_domain: "seecs.edu.pk"`)
- [ ] Verification badges on frontend (show domain when email-verified)
- [ ] Filter: "verified developers only" on leaderboard and lists
- [ ] Admin UI: manage allowed domains, re-run email checks, manual override

### Allowed Email Domains (Default)

| Domain | Affiliation |
|--------|-------------|
| `nust.edu.pk` | NUST (main) |
| `seecs.edu.pk` | SEECS — School of Electrical Engineering & Computer Science |

Additional school/department domains can be added via admin settings.

### Verification Statuses

| Status | Meaning |
|--------|---------|
| `registered` | Added with username (+ optional email); not yet verified |
| `email_verified` | Provided email matches an allowed NUST-affiliated domain |
| `manual_verified` | Admin manually confirmed |

### Exit Criteria

- [ ] Developers marked `email_verified` when their stored email matches an allowed domain
- [ ] Verification badges visible on profiles and leaderboard
- [ ] "Verified only" filter works on public pages
- [ ] Admin can add/remove allowed domains and override status

### Admin API Endpoints

| Method | Path | Auth |
|--------|------|------|
| GET | `/api/v1/admin/verification/domains` | Admin |
| POST | `/api/v1/admin/verification/domains` | Admin |
| DELETE | `/api/v1/admin/verification/domains/:domain` | Admin |
| POST | `/api/v1/admin/verification/run` | Admin |

---

## Milestone 7 — Production Hardening

**Target:** Week 8–9  
**Goal:** Deployable, monitored, documented.

### Tasks

- [x] Production Dockerfile (multi-stage Go build + Next.js standalone)
- [x] Docker Compose: PostgreSQL, Redis, API, worker, web
- [ ] Environment-specific config (dev / staging / prod)
- [x] Rate limit monitoring in admin dashboard
- [ ] Error alerting hooks (log aggregation ready)
- [x] Database indexes for leaderboard and snapshot queries
- [x] API response caching strategy (Redis TTLs per endpoint)
- [x] Comprehensive README: architecture, API docs, deployment guide
- [ ] Seed script for demo data
- [ ] GitHub Actions CI (lint, test, build)

### Exit Criteria

- [x] `docker compose up` starts all services (with documented port/env caveats)
- [ ] Platform deployable to staging/prod with env-specific config
- [x] Another developer can run locally using README alone
- [x] Leaderboard and stats queries perform well at 15+ developers (50+ untested at scale)
- [x] Admin can monitor GitHub API quota usage

---

## Milestone 8 — Future Enhancements (Post-v1 Backlog)

**Goal:** Track ideas without blocking v1 launch.

### Claim / Join Flow (partially done)

- [x] **Join request** — user submits GitHub username at `/join` (optional email, display name, message)
- [x] **Admin approval flow** — admin approves/rejects; approve creates developer + enqueues sync
- [x] **Duplicate prevention** — no repeated GitHub profiles (registered or pending)
- [ ] **Developer self-registration** — instant add without admin (deferred)
- [ ] **GitHub OAuth login** — developers manage their own profile after claiming

### Other Backlog

- [ ] PR/issue stats via GraphQL (per-repo)
- [ ] Cohort/year filters (batch of NUST graduates)
- [ ] Export stats (CSV/JSON)
- [ ] Compare two developers side-by-side
- [ ] Monthly "NUST Dev of the Month" automated ranking
- [ ] Public API rate limiting for third-party consumers
- [ ] Leaderboard sparklines and rank trend indicators
- [ ] Stars/forks growth charts from snapshots
- [ ] Dev card image export (PNG download)
- [ ] OpenAPI / Swagger documentation

---

## v1 Success Metrics

| Metric | Target | Current |
|--------|--------|---------|
| Registered & synced developers | 50+ | 🔄 ~15 synced |
| Public page load time | < 2s | ✅ |
| Sync reliability | No 429 errors within rate limits | ✅ (with PAT + worker) |
| Live chart types | 6+ | ✅ (heatmap, languages, activity, innovation trends, OSS, leaderboard) |
| Admin workflow | Add username / sync / view stats | ✅ |
| Join request workflow | Submit → admin approve → sync | ✅ |

---

## Risk Register

| Risk | Impact | Mitigation |
|------|--------|------------|
| GitHub rate limits with many developers | Sync stalls | Staggered sync, Redis queue, PAT, aggressive persistence |
| Events API 90-day limit | Incomplete history | GraphQL contribution calendar + daily snapshots |
| Private repos invisible | Incomplete stats | Document clearly; only public data |
| Fake NUST affiliation | Bad data | Email domain allowlist (M6) + admin manual override + join review |
| GitHub API changes | Broken sync | Version-pin API calls; abstract client interface |
| Large repo lists per user | Slow sync | Paginate; cap language fetch to top 20 repos by stars |
| Stale API server after deploy | New routes 404 | Restart `cmd/server` after backend changes |

---

## Implementation Order

**v1 launch path** — email verification and OAuth self-registration are explicitly deferred:

```
M0 Foundation
  └─► M1 Registry (username + optional email + join requests)
        └─► M2 GitHub Client
              └─► M3 Sync Worker
                    └─► M4 Stats API (+ innovation / OSS / activity)
                          └─► M5 Dashboard + M5b Revamp  ← usable public platform
                                └─► M7 Production  ← v1 launch
                                      └─► M6 Email verification (optional, later)
                                            └─► M8 OAuth self-registration (last priority)
```

Work strictly milestone-by-milestone for M0–M7. M6 and full M8 self-registration do not block v1 launch.

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
│   ├── githubutil/
│   ├── models/
│   ├── repository/
│   ├── service/
│   ├── handler/
│   ├── cache/
│   └── worker/
├── migrations/
├── web/
│   ├── app/
│   ├── components/
│   └── lib/api.ts
├── docker-compose.yml
├── milestones.md
├── web/REVAMP_PLAN.md
├── go.mod
└── README.md
```

---

*Last updated: June 2026*
