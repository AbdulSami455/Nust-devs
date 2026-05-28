"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, type Developer, type Overview, type PublicRepo } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function HomePage() {
  const [overview, setOverview] = useState<Overview | null>(null);
  const [topDevs, setTopDevs] = useState<Developer[]>([]);
  const [topProjects, setTopProjects] = useState<PublicRepo[]>([]);

  useEffect(() => {
    api.public.overview().then(setOverview).catch(() => {});
    api.public.developers.list(1, 6).then(setTopDevs).catch(() => {});
    api.public.topProjects().then((p) => setTopProjects(p.slice(0, 6))).catch(() => {});
  }, []);

  return (
    <div className="min-h-screen bg-muted/40">
      <header className="border-b bg-background px-6 py-4 flex items-center justify-between">
        <h1 className="text-xl font-bold">NUST Devs</h1>
        <nav className="flex gap-4 text-sm">
          <Link href="/developers" className="hover:underline">Developers</Link>
          <Link href="/leaderboard" className="hover:underline">Leaderboard</Link>
          <Link href="/admin" className="hover:underline text-muted-foreground">Admin</Link>
        </nav>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-10 space-y-10">
        <div className="space-y-1">
          <h2 className="text-3xl font-bold">NUST Developer Activity</h2>
          <p className="text-muted-foreground">Track contributions, repos, and stats from NUST developers on GitHub.</p>
        </div>

        {/* Overview stats */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          {[
            { label: "Developers", value: overview?.total_developers ?? "—" },
            { label: "Repositories", value: overview?.total_repos ?? "—" },
            { label: "Total Stars", value: overview?.total_stars ?? "—" },
            { label: "Contributions", value: overview?.total_contributions?.toLocaleString() ?? "—" },
          ].map(({ label, value }) => (
            <Card key={label}>
              <CardHeader className="pb-1">
                <CardTitle className="text-sm font-medium text-muted-foreground">{label}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{value}</p>
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Top developers */}
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Top Developers</h3>
            <Link href="/developers" className="text-sm text-muted-foreground hover:underline">View all →</Link>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {topDevs.map((dev) => (
              <Link key={dev.id} href={`/developers/${dev.github_username}`}>
                <Card className="hover:border-foreground/30 transition-colors cursor-pointer h-full">
                  <CardContent className="pt-4 flex gap-3 items-start">
                    {dev.avatar_url && (
                      <img src={dev.avatar_url} alt={dev.github_username} className="w-10 h-10 rounded-full" />
                    )}
                    <div className="min-w-0 flex-1">
                      <p className="font-medium truncate">{dev.display_name ?? dev.github_username}</p>
                      <p className="text-sm text-muted-foreground font-mono truncate">@{dev.github_username}</p>
                      <div className="flex gap-3 mt-2 text-xs text-muted-foreground">
                        <span>★ {dev.total_stars}</span>
                        <span>{dev.public_repos} repos</span>
                        <span>{dev.followers} followers</span>
                      </div>
                    </div>
                    {dev.verification_status === "email_verified" && (
                      <Badge variant="default" className="shrink-0 text-xs">verified</Badge>
                    )}
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        </section>

        {/* Top projects */}
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Top Projects</h3>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {topProjects.map((repo) => (
              <a key={repo.id} href={repo.url} target="_blank" rel="noopener noreferrer">
                <Card className="hover:border-foreground/30 transition-colors cursor-pointer h-full">
                  <CardContent className="pt-4 space-y-1">
                    <p className="font-mono font-medium text-sm truncate">{repo.full_name}</p>
                    <p className="text-xs text-muted-foreground line-clamp-2">{repo.description || "No description"}</p>
                    <div className="flex gap-3 pt-1 text-xs text-muted-foreground">
                      <span>★ {repo.stars}</span>
                      {repo.language && <Badge variant="secondary" className="text-xs">{repo.language}</Badge>}
                    </div>
                  </CardContent>
                </Card>
              </a>
            ))}
          </div>
        </section>
      </main>
    </div>
  );
}
