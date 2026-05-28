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
  verification_status: string;
  last_synced_at?: string;
  created_at: string;
}

export interface PublicRepo {
  id: string;
  name: string;
  full_name: string;
  description: string;
  url: string;
  language?: string;
  stars: number;
  forks: number;
  is_fork: boolean;
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

export interface LanguageStat {
  language: string;
  bytes: number;
  repo_count: number;
}

export const api = {
  login: (email: string, password: string) =>
    request<{ token: string }>("/api/v1/admin/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  // Admin endpoints
  admin: {
    developers: {
      list: () => request<Developer[]>("/api/v1/admin/developers"),
      create: (data: { github_username: string; email: string; display_name?: string; notes?: string }) =>
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
  },

  // Public endpoints (no auth)
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
    leaderboard: (sortBy = "activity_score", page = 1, limit = 20) =>
      request<Developer[]>(`/api/v1/leaderboard?sort_by=${sortBy}&page=${page}&limit=${limit}`),
    topProjects: () => request<PublicRepo[]>("/api/v1/projects/top"),
    overview: () => request<Overview>("/api/v1/stats/overview"),
    languages: () => request<LanguageStat[]>("/api/v1/stats/languages"),
  },

  // Keep old shape for backward compat with dashboard page
  developers: {
    list: () => request<Developer[]>("/api/v1/admin/developers"),
    create: (data: { github_username: string; email: string; display_name?: string; notes?: string }) =>
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
