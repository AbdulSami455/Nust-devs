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

## AI Roadmap

These are intentionally lightweight AI features that are useful, low-risk, and easy to ship.
Prioritization should favor profile, leaderboard, compare, and admin workflows first.

### AI Enhancements to Existing Features

| # | Enhancement | Existing Surface |
|---|-------------|------------------|
| 1 | AI-generated developer summary card from GitHub bio, repos, and stats | Developer Profile |
| 2 | AI-written key strengths on profile pages | Developer Profile |
| 3 | AI-written project summary for each repo card | Projects |
| 4 | AI-generated "why this rank" explanation for leaderboard entries | Leaderboard |
| 5 | AI comparison summary for two developers | Developer Profile, Leaderboard |
| 6 | AI-generated home dashboard digest with the latest sync changes | Home |
| 7 | AI narration for charts like spikes, growth, and streaks | Stats, Innovation |
| 8 | AI-written search suggestions when a query returns no results | Developers, Projects |
| 9 | AI-generated SEO title and description for profile pages | Public Pages |
| 10 | ✅ AI-written one-line insights for top projects and community trends | Projects, Home, Stats |
| 11 | ✅ AI join-request summary for admins | Join, Admin |
| 12 | ✅ AI duplicate warning when a request matches an existing profile closely | Join, Admin |
| 13 | ✅ AI summary of recent sync changes for admins | Admin |
| 14 | ✅ AI-generated weekly community report | Home, Stats |
| 15 | ✅ AI suggested tags for projects and profiles | Projects, Developers |
| 16 | ✅ AI quick explanation of score breakdowns on profile pages | Developer Profile |
| 17 | ✅ AI-generated homepage "what changed today" banner | Home |
| 18 | ✅ AI-written share text for profile and project links | Public Pages |
| 19 | ✅ AI-generated "top achievements" bullets from a developer's activity | Developer Profile |
| 20 | ✅ AI-written recent activity recap for profile pages | Developer Profile |
| 21 | ✅ AI-suggested completion tips for incomplete profiles | Developer Profile, Admin |
| 22 | ✅ AI normalizes skill and language names into consistent tags | Developers, Projects |
| 23 | ✅ AI-generated explanation of new badges or rank changes | Leaderboard, Profile |
| 24 | ✅ AI one-paragraph project impact summary for featured repos | Projects |

### New AI-Heavy Features

| # | Feature | Scope |
|---|---------|-------|
| 1 | Ask the platform simple questions in chat, like "Who are the top Python developers?" | New |
| 2 | Auto-fill developer summaries from the latest sync data | New |
| 3 | Generate weekly digest emails or in-app summaries | New |
| 4 | Simple AI search assistant that turns natural language into filters | Developers, Projects |
| 5 | Admin helper that summarizes pending join requests | Admin |
| 6 | Admin helper that flags possible duplicate profiles | Admin |
| 7 | Simple "recommend similar developers" suggestions | Developer Profile |
| 8 | Simple "recommend similar projects" suggestions | Projects |
| 9 | AI form autofill hints for join requests and admin notes | Join, Admin |
| 10 | AI-generated "next best action" suggestion for admins reviewing requests | Admin |
| 11 | AI compact digest for leaderboard changes each week | Leaderboard |
| 12 | AI summary of a developer's top repos and languages | Developer Profile |
| 13 | AI compare wizard that explains one developer's strengths vs another | Compare, Developer Profile |
| 14 | AI profile completeness coach that tells users what to add next | Developer Profile |
| 15 | AI natural-language filter assistant for finding developers and projects | Developers, Projects |
| 16 | AI collaboration matcher that suggests developers with similar interests | Developer Profile, Projects |
| 17 | AI-generated FAQ answers for common profile and join questions | Join, Profile |
| 18 | AI-powered notification explainer that translates alerts into plain language | Home, Admin |

---

## v1 Scope

The v1 launch path is:

M0 Foundation -> M1 Registry -> M2 GitHub Client -> M3 Sync Worker -> M4 Stats API -> M5 Dashboard -> M5b Revamp -> M7 Production

M6 email verification and full M8 self-registration remain out of the v1 critical path.
