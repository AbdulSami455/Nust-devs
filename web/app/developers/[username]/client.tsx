"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  api,
  type Developer,
  type PublicRepo,
  type ContributionDay,
  type ContributionStats,
} from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { buttonVariants } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { DevCard } from "@/components/profile/dev-card";
import { ContributionHeatmap } from "@/components/profile/contribution-heatmap";
import { ContributionStatsPanel } from "@/components/profile/contribution-stats";
import { DeveloperSummaryCard } from "@/components/ai/developer-summary";
import { ProfileCompletenessCard } from "@/components/ai/profile-completeness";
import { cn } from "@/lib/utils";

export function ProfileClient({ username }: { username: string }) {
  const [dev, setDev] = useState<Developer | null>(null);
  const [repos, setRepos] = useState<PublicRepo[]>([]);
  const [contributions, setContributions] = useState<ContributionDay[]>([]);
  const [contributionStats, setContributionStats] = useState<ContributionStats | null>(null);
  const [developers, setDevelopers] = useState<Developer[]>([]);
  const [compareWith, setCompareWith] = useState("");
  const [rank, setRank] = useState<number | undefined>();
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    Promise.all([
      api.public.developers.get(username),
      api.public.developers.repos(username),
      api.public.developers.contributions(username),
      api.public.developers.contributionStats(username).catch(() => null),
      api.public.leaderboard("activity_score", 1, 100),
      api.public.developers.list(1, 100),
    ])
      .then(([d, r, c, stats, board, devs]) => {
        setDev(d);
        setRepos(r ?? []);
        setContributions(c ?? []);
        setContributionStats(stats);
        setDevelopers((devs ?? []).filter((item) => item.github_username !== username));
        setCompareWith("");
        const idx = board.findIndex((x) => x.github_username === username);
        setRank(idx >= 0 ? idx + 1 : undefined);
      })
      .catch(() => setNotFound(true))
      .finally(() => setLoading(false));
  }, [username]);

  if (loading) {
    return (
      <div className="mx-auto max-w-5xl space-y-6 px-4 py-8 sm:px-6">
        <Skeleton className="h-64 rounded-2xl" />
        <Skeleton className="h-32 rounded-2xl" />
        <Skeleton className="h-48 rounded-2xl" />
      </div>
    );
  }

  if (notFound || !dev) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-24 text-center sm:px-6">
        <h1 className="text-xl font-semibold">Developer not found</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          This profile is not tracked yet or the username is incorrect.
        </p>
        <Link href="/developers" className="mt-4 inline-block text-sm text-primary hover:underline">
          Browse developers
        </Link>
      </div>
    );
  }

  const originalRepos = repos.filter((r) => !r.is_fork);
  const topRepos = [...repos].sort((a, b) => b.stars - a.stars).slice(0, 6);

  return (
    <div className="mx-auto max-w-5xl space-y-10 px-4 py-8 sm:px-6">
      <DevCard dev={dev} rank={rank} />

      <Card className="border-primary/15 bg-gradient-to-br from-primary/5 via-card to-card">
        <CardHeader>
          <CardTitle>Compare this developer</CardTitle>
          <CardDescription>
            Pick a second developer to open a side-by-side AI comparison.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form
            className="flex flex-col gap-3 sm:flex-row sm:items-end"
            onSubmit={(e) => e.preventDefault()}
          >
            <label className="flex-1 space-y-2">
              <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                Second developer
              </span>
              <select
                value={compareWith}
                onChange={(e) => setCompareWith(e.target.value)}
                className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
              >
                <option value="">Select developer</option>
                {developers.map((candidate) => (
                  <option key={candidate.id} value={candidate.github_username}>
                    {candidate.display_name ?? candidate.github_username} (@{candidate.github_username})
                  </option>
                ))}
              </select>
            </label>

            <Link
              href={compareWith ? `/compare?left=${username}&right=${compareWith}` : "#"}
              aria-disabled={!compareWith}
              className={cn(
                buttonVariants(),
                "sm:w-auto",
                !compareWith && "pointer-events-none opacity-50"
              )}
            >
              Compare now
            </Link>
          </form>
          <p className="mt-3 text-xs text-muted-foreground">
            The current profile stays fixed on the left so you can compare it with another tracked developer.
          </p>
        </CardContent>
      </Card>

      <DeveloperSummaryCard username={username} />

      <ProfileCompletenessCard dev={dev} repos={repos} />

      {contributions.length > 0 && (
        <div className="bento-card">
          <ContributionHeatmap days={contributions} />
        </div>
      )}

      {contributionStats && <ContributionStatsPanel stats={contributionStats} />}

      <section className="space-y-4">
        <div className="flex items-end justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Top repositories</h2>
            <p className="text-sm text-muted-foreground">
              {originalRepos.length} original · {repos.length - originalRepos.length} forks
            </p>
          </div>
        </div>

        {topRepos.length === 0 ? (
          <p className="text-sm text-muted-foreground">No repositories synced yet.</p>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {topRepos.map((repo) => (
              <a
                key={repo.id}
                href={repo.url}
                target="_blank"
                rel="noopener noreferrer"
                className="bento-card block transition-colors hover:border-primary/40"
              >
                <div className="flex items-start justify-between gap-2">
                  <p className="font-mono text-sm font-medium">{repo.name}</p>
                  {repo.is_fork ? (
                    <Badge variant="secondary" className="text-[10px]">
                      Fork
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="text-[10px]">
                      OSS
                    </Badge>
                  )}
                </div>
                <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">
                  {repo.description || "No description"}
                </p>
                <div className="mt-3 flex gap-3 text-xs text-muted-foreground">
                  <span>{repo.stars} stars</span>
                  <span>{repo.forks} forks</span>
                  {repo.language && <Badge variant="outline">{repo.language}</Badge>}
                </div>
              </a>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
