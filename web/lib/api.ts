const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `HTTP ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export interface Developer {
  id: string;
  github_username: string;
  email?: string;
  display_name?: string;
  avatar_url?: string;
  bio?: string;
  location?: string;
  company?: string;
  website?: string;
  public_repos: number;
  readme_repos?: number;
  total_stars: number;
  followers: number;
  following: number;
  activity_score: number;
  builder_score: number;
  contributor_score: number;
  reviewer_score: number;
  community_score: number;
  pr_contributions: number;
  issue_contributions: number;
  review_contributions: number;
  contribution_period_start?: string;
  contribution_period_end?: string;
  current_streak?: number;
  longest_streak?: number;
  streak_multiplier?: number;
  xp?: number;
  power_level?: number;
  verification_status: string;
  last_synced_at?: string;
  created_at: string;
}

export interface SparkPoint {
  date: string;
  value: number;
}

export interface LeaderboardEntry extends Developer {
  rank: number;
  rank_delta_7d?: number | null;
  rank_delta_30d?: number | null;
  score_delta_7d?: number | null;
  score_delta_30d?: number | null;
  sparkline?: SparkPoint[];
}

export interface PublicRepo {
  id: string;
  name: string;
  full_name: string;
  owner?: string;
  description: string;
  url: string;
  language?: string;
  stars: number;
  forks: number;
  is_fork: boolean;
  pushed_at?: string;
  stars_growth_30d?: number | null;
  forks_growth_30d?: number | null;
  sparkline?: SparkPoint[];
}

export interface ProjectSummary {
  repo_id: string;
  headline: string;
  summary: string;
  model_version: string;
  generated_at: string;
}

export interface RankInsight {
  developer_id: string;
  headline: string;
  summary: string;
  highlights: string[];
  model_version: string;
  generated_at: string;
}

export interface NormalizedTags {
  entity_type: string;
  entity_id: string;
  headline: string;
  summary: string;
  languages: string[];
  skills: string[];
  tags: string[];
  model_version: string;
  generated_at: string;
}

export interface ContributionDay {
  date: string;
  count: number;
}

export interface RepoContributionStat {
  repo_full_name: string;
  repo_url: string;
  pull_requests: number;
  issues: number;
  reviews: number;
  total: number;
}

export interface ContributionStats {
  period_start: string;
  period_end: string;
  pull_requests: number;
  issues: number;
  reviews: number;
  by_repository: RepoContributionStat[];
}

export interface StreakSummary {
  devs_on_7plus_streak: number;
  devs_on_30plus_streak: number;
  longest_active_streak: number;
}

export interface DevOfMonthWinner {
  year: number;
  month: number;
  score: number;
  activity_points: number;
  rank_gain: number;
  stars_gained: number;
  power_title?: string;
  developer: Developer;
}

export interface WrappedReport {
  year: number;
  username: string;
  display_name?: string;
  avatar_url?: string;
  total_contributions: number;
  top_repo?: string;
  top_repo_stars: number;
  rank_start: number;
  rank_end: number;
  rank_change: number;
  activity_percentile: number;
  top_languages: NameCount[];
  power_level: number;
  power_title: string;
  xp: number;
  current_streak: number;
  longest_streak: number;
  pr_contributions: number;
  total_stars: number;
  public_repos: number;
  highlights: string[];
}

export interface Overview {
  total_developers: number;
  total_repos: number;
  total_stars: number;
  total_contributions: number;
}

export interface CommunityActivityDay {
  date: string;
  count: number;
}

export interface LanguageStat {
  language: string;
  bytes: number;
  repo_count: number;
}

export interface ActivityEvent {
  type: string;
  username: string;
  repo?: string;
  message: string;
  occurred_at: string;
}

export interface InnovationGraph {
  granularity: string;
  pushes: TrendPoint[];
  repositories: TrendPoint[];
  developers: TrendPoint[];
  organizations: TrendPoint[];
  net_new_stars: TrendPoint[];
  languages: NameCount[];
  licenses: NameCount[];
  top_organizations: NameCount[];
  top_contributors: ContributorStat[];
}

export interface TrendPoint {
  period: string;
  label: string;
  value: number;
}

export interface NameCount {
  name: string;
  count: number;
}

export interface ContributorStat {
  username: string;
  name?: string;
  score: number;
  stars: number;
}

export interface DeveloperRequest {
  id: string;
  github_username: string;
  email?: string;
  display_name?: string;
  batch?: string;
  course?: string;
  message?: string;
  status: "pending" | "approved" | "rejected";
  admin_notes?: string;
  reviewed_at?: string;
  created_at: string;
}

export interface UsernameCheck {
  available: boolean;
  reason?: string;
  username?: string;
}

export interface OSSStats {
  original_projects: number;
  fork_projects: number;
  total_stars: number;
  total_forks_received: number;
  contributors: number;
  top_language?: string;
}

export interface AuditLog {
  id: string;
  actor_type: string;
  actor_id?: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  method?: string;
  path?: string;
  status_code: number;
  ip?: string;
  user_agent?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface AgentRun {
  id: string;
  session_id: string;
  agent_name: string;
  user_message: string;
  input_hash: string;
  status: string;
  ip?: string;
  user_agent?: string;
  tool_calls: number;
  latency_ms: number;
  error_message?: string;
  response_chars: number;
  created_at: string;
  finished_at?: string;
}

export interface AgentRunEvent {
  id: string;
  run_id: string;
  event_type: string;
  tool_name?: string;
  payload?: Record<string, unknown>;
  latency_ms: number;
  success: boolean;
  created_at: string;
}

export interface ObservabilityOverview {
  total_audit_logs: number;
  agent_runs_24h: number;
  agent_success_rate_24h: number;
  avg_agent_latency_ms: number;
  active_agent_runs: number;
  last_agent_run_at?: string;
}

export interface ObservabilityResponse {
  overview: ObservabilityOverview;
  recent_logs: AuditLog[];
  recent_runs: AgentRun[];
  recent_events: AgentRunEvent[];
}

export type ProjectCategory = "all" | "original" | "forks";
export type ProjectSort = "stars" | "recent" | "forks" | "growth";

export const api = {
  login: (email: string, password: string) =>
    request<{ ok: boolean }>("/api/v1/admin/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  logout: () =>
    request<{ ok: boolean }>("/api/v1/admin/auth/logout", {
      method: "POST",
    }),

  ai: {
    compareDevelopers: (left: string, right: string) =>
      request<DeveloperComparison>(
        `/api/v1/ai/compare?left=${encodeURIComponent(left)}&right=${encodeURIComponent(right)}`
      ),
  },

  admin: {
    developers: {
      list: () => request<Developer[]>("/api/v1/admin/developers"),
      create: (data: { github_username: string; email?: string; display_name?: string; notes?: string }) =>
        request<Developer>("/api/v1/admin/developers", {
          method: "POST",
          body: JSON.stringify(data),
        }),
      update: (id: string, data: { email?: string; display_name?: string; notes?: string }) =>
        request<Developer>(`/api/v1/admin/developers/${id}`, {
          method: "PATCH",
          body: JSON.stringify(data),
        }),
      delete: (id: string) =>
        request<void>(`/api/v1/admin/developers/${id}`, { method: "DELETE" }),
    },
    sync: {
      trigger: (id?: string) =>
        request<unknown>(`/api/v1/admin/sync${id ? `?id=${id}` : ""}`, { method: "POST" }),
      status: () => request<unknown>("/api/v1/admin/sync/status"),
    },
    profileRequests: {
      list: (status?: string) =>
        request<DeveloperRequest[]>(
          `/api/v1/admin/profile-requests${status ? `?status=${status}` : ""}`
        ),
      approve: (id: string, admin_notes?: string) =>
        request<{ request: DeveloperRequest; developer: Developer }>(
          `/api/v1/admin/profile-requests/${id}/approve`,
          {
            method: "POST",
            body: JSON.stringify(admin_notes ? { admin_notes } : {}),
          }
        ),
      reject: (id: string, admin_notes?: string) =>
        request<DeveloperRequest>(`/api/v1/admin/profile-requests/${id}/reject`, {
          method: "POST",
          body: JSON.stringify(admin_notes ? { admin_notes } : {}),
        }),
    },
    observability: {
      overview: () => request<ObservabilityResponse>("/api/v1/admin/observability"),
      logs: (limit = 50) =>
        request<AuditLog[]>(`/api/v1/admin/observability/logs?limit=${limit}`),
      agentRuns: (limit = 25) =>
        request<AgentRun[]>(`/api/v1/admin/observability/agent-runs?limit=${limit}`),
      agentEvents: (limit = 40) =>
        request<AgentRunEvent[]>(`/api/v1/admin/observability/agent-events?limit=${limit}`),
    },
  },

  public: {
    developers: {
      list: (page = 1, limit = 20) =>
        request<Developer[]>(`/api/v1/developers?page=${page}&limit=${limit}`),
      get: (username: string) =>
        request<Developer>(`/api/v1/developers/${username}`),
      repos: (username: string) =>
        request<PublicRepo[]>(`/api/v1/developers/${username}/repos`),
      contributions: (username: string) =>
        request<ContributionDay[]>(`/api/v1/developers/${username}/contributions`),
      contributionStats: (username: string) =>
        request<ContributionStats>(`/api/v1/developers/${username}/contribution-stats`),
      wrapped: (username: string, year?: number) =>
        request<WrappedReport>(
          `/api/v1/developers/${username}/wrapped${year ? `?year=${year}` : ""}`
        ),
    },
    leaderboard: (
      sortBy = "activity_score",
      page = 1,
      limit = 20,
      opts?: { view?: "default" | "rising"; period?: 7 | 30 }
    ) => {
      const params = new URLSearchParams({
        sort_by: sortBy,
        page: String(page),
        limit: String(limit),
      });
      if (opts?.view === "rising") {
        params.set("view", "rising");
        params.set("period", String(opts.period ?? 7));
      }
      return request<LeaderboardEntry[]>(`/api/v1/leaderboard?${params}`);
    },
    topProjects: (opts?: {
      category?: ProjectCategory;
      language?: string;
      sort?: ProjectSort;
      limit?: number;
    }) => {
      const params = new URLSearchParams();
      if (opts?.category && opts.category !== "all") params.set("category", opts.category);
      if (opts?.language) params.set("language", opts.language);
      if (opts?.sort) params.set("sort", opts.sort);
      if (opts?.limit) params.set("limit", String(opts.limit));
      const q = params.toString();
      return request<PublicRepo[]>(`/api/v1/projects/top${q ? `?${q}` : ""}`);
    },
    fastestGrowingProjects: (days = 30, limit = 8) =>
      request<PublicRepo[]>(
        `/api/v1/projects/fastest-growing?days=${days}&limit=${limit}`
      ),
    repoGrowth: (repoId: string, days = 30) =>
      request<SparkPoint[]>(`/api/v1/repos/${repoId}/growth?days=${days}`),
    overview: () => request<Overview>("/api/v1/stats/overview"),
    languages: () => request<LanguageStat[]>("/api/v1/stats/languages"),
    communityActivity: (days = 30) =>
      request<CommunityActivityDay[]>(`/api/v1/stats/community-activity?days=${days}`),
    spotlight: () => request<Developer>("/api/v1/developers/spotlight"),
    recentActivity: (limit = 15) =>
      request<ActivityEvent[]>(`/api/v1/activity/recent?limit=${limit}`),
    openSource: () => request<OSSStats>("/api/v1/stats/open-source"),
    innovationGraph: (granularity = "quarterly", periods = 8) =>
      request<InnovationGraph>(
        `/api/v1/stats/innovation-graph?granularity=${granularity}&periods=${periods}`
      ),
    streakSummary: () => request<StreakSummary>("/api/v1/stats/streak-summary"),
    devOfMonth: (limit = 12) =>
      request<DevOfMonthWinner[]>(`/api/v1/dev-of-month?limit=${limit}`),
    submitProfileRequest: (data: {
      github_username: string;
      email?: string;
      display_name?: string;
      batch?: string;
      course?: string;
      message?: string;
    }) =>
      request<DeveloperRequest>("/api/v1/profile-requests", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    checkProfileUsername: (username: string) =>
      request<UsernameCheck>(
        `/api/v1/profile-requests/check?username=${encodeURIComponent(username)}`
      ),
  },

  developers: {
    list: () => request<Developer[]>("/api/v1/admin/developers"),
    create: (data: { github_username: string; email?: string; display_name?: string; notes?: string }) =>
      request<Developer>("/api/v1/admin/developers", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    update: (id: string, data: { email?: string; display_name?: string; notes?: string }) =>
      request<Developer>(`/api/v1/admin/developers/${id}`, {
        method: "PATCH",
        body: JSON.stringify(data),
      }),
    delete: (id: string) =>
      request<void>(`/api/v1/admin/developers/${id}`, { method: "DELETE" }),
  },
};

// ── AI ────────────────────────────────────────────────────────────────────────

export interface DeveloperSummary {
  developer_id: string;
  headline: string;
  summary: string;
  strengths: string[];
  model_version: string;
  generated_at: string;
}

export interface DeveloperComparison {
  left: Developer;
  right: Developer;
  left_rank?: number;
  right_rank?: number;
  headline: string;
  summary: string;
  takeaways: string[];
  shared_strengths: string[];
  verdict: string;
  source: "ai" | "fallback";
  model_version: string;
  generated_at: string;
}

export interface ChatMessage {
  role: "user" | "assistant";
  content: string;
}

export interface ChatAgentEvent {
  type: "status" | "tool_call" | "tool_done";
  message?: string;
  tool_name?: string;
  success?: boolean;
  latency_ms?: number;
}

/** Fetch the cached AI summary for a developer. Returns null if unavailable. */
export async function fetchDeveloperSummary(username: string): Promise<DeveloperSummary | null> {
  try {
    return await request<DeveloperSummary>(`/api/v1/developers/${username}/summary`);
  } catch {
    return null;
  }
}

/** Fetch the cached AI summary for a featured project. Returns null if unavailable. */
export async function fetchProjectSummary(repoId: string): Promise<ProjectSummary | null> {
  try {
    return await request<ProjectSummary>(`/api/v1/repos/${repoId}/summary`);
  } catch {
    return null;
  }
}

/** Fetch the cached AI rank/badge insight for a developer. Returns null if unavailable. */
export async function fetchRankInsight(username: string): Promise<RankInsight | null> {
  try {
    return await request<RankInsight>(`/api/v1/developers/${username}/rank-insight`);
  } catch {
    return null;
  }
}

/** Fetch normalized language/skill tags for a developer. Returns null if unavailable. */
export async function fetchDeveloperNormalizedTags(username: string): Promise<NormalizedTags | null> {
  try {
    return await request<NormalizedTags>(`/api/v1/developers/${username}/normalized-tags`);
  } catch {
    return null;
  }
}

/** Fetch normalized language/skill tags for a project. Returns null if unavailable. */
export async function fetchProjectNormalizedTags(repoId: string): Promise<NormalizedTags | null> {
  try {
    return await request<NormalizedTags>(`/api/v1/repos/${repoId}/normalized-tags`);
  } catch {
    return null;
  }
}

/**
 * Send a chat message and stream the response via SSE.
 * Calls onToken for each text chunk, onDone when finished, onError on failure.
 * Returns an AbortController so the caller can cancel.
 */
export function streamChat(
  message: string,
  history: ChatMessage[],
  onToken: (token: string) => void,
  onAgentEvent: (event: ChatAgentEvent) => void,
  onDone: () => void,
  onError: (err: string) => void,
): AbortController {
  const controller = new AbortController();

  (async () => {
    let res: Response;
    try {
      res = await fetch(`${BASE}/api/v1/ai/chat`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message, history }),
        signal: controller.signal,
      });
    } catch (e: unknown) {
      if ((e as { name?: string }).name !== "AbortError") onError("Connection failed");
      return;
    }

    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      onError(body.error ?? `Error ${res.status}`);
      return;
    }

    const reader = res.body?.getReader();
    if (!reader) { onError("No response stream"); return; }

    const decoder = new TextDecoder();
    let buffer = "";

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop() ?? "";
      let currentEvent = "message";
      for (const line of lines) {
        if (line.startsWith("event: ")) {
          currentEvent = line.slice(7).trim();
          continue;
        }
        if (line.startsWith("data: ")) {
          const payload = line.slice(6).trim();
          if (currentEvent === "done") {
            onDone();
            return;
          }
          if (currentEvent === "token") {
            try {
              onToken(JSON.parse(payload) as string);
            } catch {
              onToken(payload);
            }
            continue;
          }
          if (currentEvent === "status" || currentEvent === "tool_call" || currentEvent === "tool_done") {
            try {
              onAgentEvent(JSON.parse(payload) as ChatAgentEvent);
            } catch {
              onAgentEvent({ type: "status", message: payload });
            }
            continue;
          }
        }
      }
    }
    onDone();
  })();

  return controller;
}
