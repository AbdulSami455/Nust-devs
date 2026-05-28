"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, type Developer } from "@/lib/api";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

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
    <div className="min-h-screen bg-muted/40">
      <header className="border-b bg-background px-6 py-4 flex items-center justify-between">
        <Link href="/" className="text-xl font-bold">NUST Devs</Link>
        <nav className="flex gap-4 text-sm">
          <Link href="/developers" className="hover:underline font-medium">Developers</Link>
          <Link href="/leaderboard" className="hover:underline">Leaderboard</Link>
        </nav>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-8 space-y-6">
        <div className="flex items-center justify-between gap-4">
          <h2 className="text-2xl font-bold">Developers</h2>
          <Input
            placeholder="Search by name or location…"
            className="max-w-xs"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>

        {loading ? (
          <p className="text-muted-foreground">Loading…</p>
        ) : filtered.length === 0 ? (
          <p className="text-muted-foreground">No developers found.</p>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {filtered.map((dev) => (
              <Link key={dev.id} href={`/developers/${dev.github_username}`}>
                <Card className="hover:border-foreground/30 transition-colors cursor-pointer h-full">
                  <CardContent className="pt-4 flex gap-3 items-start">
                    {dev.avatar_url ? (
                      <img src={dev.avatar_url} alt={dev.github_username} className="w-12 h-12 rounded-full shrink-0" />
                    ) : (
                      <div className="w-12 h-12 rounded-full bg-muted shrink-0 flex items-center justify-center text-lg font-bold text-muted-foreground">
                        {dev.github_username[0].toUpperCase()}
                      </div>
                    )}
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <p className="font-medium truncate">{dev.display_name ?? dev.github_username}</p>
                        {dev.verification_status === "email_verified" && (
                          <Badge variant="default" className="text-xs shrink-0">✓</Badge>
                        )}
                      </div>
                      <p className="text-sm text-muted-foreground font-mono">@{dev.github_username}</p>
                      {dev.location && <p className="text-xs text-muted-foreground mt-0.5">{dev.location}</p>}
                      <div className="flex gap-3 mt-2 text-xs text-muted-foreground">
                        <span>★ {dev.total_stars}</span>
                        <span>{dev.public_repos} repos</span>
                        <span>{dev.followers} followers</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
