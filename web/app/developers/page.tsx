"use client";

import { useEffect, useState } from "react";
import { api, type Developer } from "@/lib/api";
import { Input } from "@/components/ui/input";
import { DevCardMini } from "@/components/dashboard/dev-card-mini";
import { Skeleton } from "@/components/ui/skeleton";

export default function DevelopersPage() {
  const [developers, setDevelopers] = useState<Developer[]>([]);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.public.developers.list(1, 100).then(setDevelopers).catch(() => {}).finally(() => setLoading(false));
  }, []);

  const filtered = developers.filter((d) => {
    const q = query.toLowerCase();
    return (
      d.github_username.toLowerCase().includes(q) ||
      (d.display_name ?? "").toLowerCase().includes(q) ||
      (d.location ?? "").toLowerCase().includes(q)
    );
  });

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">Developers</h1>
          <p className="text-sm text-muted-foreground">All tracked NUST developers on GitHub</p>
        </div>
        <Input
          placeholder="Search by name or location…"
          className="max-w-xs"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
      </div>

      {loading ? (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-28 rounded-2xl" />
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <p className="text-muted-foreground">No developers found.</p>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((dev) => (
            <DevCardMini key={dev.id} dev={dev} />
          ))}
        </div>
      )}
    </div>
  );
}
