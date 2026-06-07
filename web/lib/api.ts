const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function getToken() {
  return typeof window !== "undefined" ? localStorage.getItem("token") : null;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken();
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
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
  total_stars: number;
  followers: number;
  following: number;
  activity_score: number;
  builder_score: number;
  contributor_score: number;
  reviewer_score: number;
  community_score: number;
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

export interface ContributionDay {
  date: string;
  count: number;
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

export type ProjectCategory = "all" | "original" | "forks";
export type ProjectSort = "stars" | "recent" | "forks" | "growth";

export const api = {
  login: (email: string, password: string) =>
    request<{ token: string }>("/api/v1/admin/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

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
    submitProfileRequest: (data: {
      github_username: string;
      email?: string;
      display_name?: string;
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
