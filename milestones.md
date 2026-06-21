# Nust Devs Platform - Milestones

> Track NUST developer activity, contributions, and top projects using the GitHub public API.

**Stack:** Go API + Next.js dashboard + PostgreSQL + Redis + Asynq workers

**Discovery model:** Admin can register developers directly with a GitHub username. NUST developers can also submit a join request at `/join` for admin approval. Email verification and full self-registration are deferred.

---

## Progress Overview

| Milestone | Name | Target | Status |
|-----------|------|--------|--------|
| M0 | Foundation | Week 1 | Done |
| M1 | Database and Developer Registry | Week 2 | Done |
| M2 | GitHub Client and Rate Limiter | Week 2-3 | Done |
| M3 | Sync Worker | Week 3-4 | Done |
| M4 | Stats Engine and Public API | Week 4-5 | Done |
| M5 | Public Dashboard | Week 5-7 | Done |
| M5b | Frontend Revamp and Innovation Graph | Post-M5 | Done |
| M6 | Email Verification | Week 8+ | Deferred |
| M7 | Production Hardening | Week 8-9 | In progress |
| M8 | Future Enhancements | Post-v1 | Partial |

---

## Milestones

### M0 - Foundation
Runnable monorepo skeleton with CI, config loading, logging, health checks, and migration tooling.

### M1 - Database and Developer Registry
Persistent developer registry with admin authentication, CRUD endpoints, duplicate prevention, and join request flow.

### M2 - GitHub Client and Rate Limiter
GitHub integration layer with REST and GraphQL clients, rate-limit handling, and test coverage around API access.

### M3 - Sync Worker
Background jobs that sync developer profiles, repositories, languages, and contribution data into the database.

### M4 - Stats Engine and Public API
Computed metrics, leaderboards, aggregation endpoints, and cached public read APIs.

### M5 - Public Dashboard
Frontend dashboard with developer browsing, profiles, rankings, project views, stats, join flow, and mobile navigation.

### M5b - Frontend Revamp and Innovation Graph
Dashboard polish and expanded innovation graph experience.

### M6 - Email Verification
Optional domain-based email verification for admin-provided emails. Deferred until after the core platform is stable.

### M7 - Production Hardening
Deployable production setup with Docker, monitoring, documentation, and environment-specific configuration.

### M8 - Future Enhancements
Backlog for self-registration, OAuth login, richer stats, exports, comparison tools, and other post-v1 ideas.

---

## v1 Scope

The v1 launch path is:

M0 Foundation -> M1 Registry -> M2 GitHub Client -> M3 Sync Worker -> M4 Stats API -> M5 Dashboard -> M5b Revamp -> M7 Production

M6 email verification and full M8 self-registration remain out of the v1 critical path.

