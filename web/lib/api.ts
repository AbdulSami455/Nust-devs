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
  email: string;
  display_name?: string;
  notes?: string;
  public_repos: number;
  total_stars: number;
  followers: number;
  activity_score: number;
  verification_status: string;
  last_synced_at?: string;
  created_at: string;
}

export const api = {
  login: (email: string, password: string) =>
    request<{ token: string }>("/api/v1/admin/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

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
