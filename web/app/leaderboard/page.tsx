"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, type Developer } from "@/lib/api";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

const SORT_OPTIONS = [
  { value: "activity_score", label: "Activity" },
  { value: "total_stars", label: "Stars" },
  { value: "public_repos", label: "Repos" },
  { value: "followers", label: "Followers" },
];

const LIMIT_TABS = [
  { value: 10, label: "Top 10" },
  { value: 50, label: "Top 50" },
  { value: 100, label: "All" },
];

const PODIUM = [
  { place: 2, ring: "ring-slate-400/60", bar: "h-20", label: "2nd" },
  { place: 1, ring: "ring-[var(--nust-gold)]/80", bar: "h-28", label: "1st" },
  { place: 3, ring: "ring-amber-700/60", bar: "h-16", label: "3rd" },
];

function sortValue(dev: Developer, sortBy: string) {
  switch (sortBy) {
    case "total_stars":
      return dev.total_stars;
    case "public_repos":
      return dev.public_repos;
    case "followers":
      return dev.followers;
    default:
      return Math.round(dev.activity_score);
  }
}

export default function LeaderboardPage() {
  const [developers, setDevelopers] = useState<Developer[]>([]);
  const [sortBy, setSortBy] = useState("activity_score");
  const [limit, setLimit] = useState(50);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.public
      .leaderboard(sortBy, 1, limit)
      .then(setDevelopers)
      .catch(() => setDevelopers([]))
      .finally(() => setLoading(false));
  }, [sortBy, limit]);

  const top3 = developers.slice(0, 3);
  const rest = developers.slice(3);

  return (
    <div className="mx-auto max-w-5xl space-y-8 px-4 py-8 sm:px-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">Leaderboard</h1>
        <p className="text-sm text-muted-foreground">
          Rankings among tracked NUST developers on GitHub.
        </p>
      </div>

      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-wrap gap-2">
          {LIMIT_TABS.map(({ value, label }) => (
            <button
              key={value}
              type="button"
              onClick={() => setLimit(value)}
              className={cn(
                "rounded-lg px-3 py-1.5 text-sm font-medium transition-colors",
                limit === value
                  ? "bg-primary text-primary-foreground"
                  : "border bg-background hover:bg-muted"
              )}
            >
              {label}
            </button>
          ))}
        </div>
        <div className="flex flex-wrap gap-2">
          {SORT_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              type="button"
              onClick={() => setSortBy(opt.value)}
              className={cn(
                "rounded-full px-3 py-1 text-xs font-medium transition-colors",
                sortBy === opt.value
                  ? "bg-primary text-primary-foreground"
                  : "bg-muted hover:bg-muted/80"
              )}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <Skeleton className="h-64 rounded-2xl" />
      ) : developers.length === 0 ? (
        <div className="rounded-2xl border border-dashed px-6 py-16 text-center text-muted-foreground">
          No developers synced yet.
        </div>
      ) : (
        <>
          {top3.length >= 3 && (
            <div className="bento-card">
              <p className="mb-6 text-center text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                Podium
              </p>
              <div className="flex items-end justify-center gap-4 sm:gap-8">
                {PODIUM.map(({ place, ring, bar, label }) => {
                  const dev = top3[place - 1];
                  if (!dev) return null;
                  return (
                    <Link
                      key={dev.id}
                      href={`/developers/${dev.github_username}`}
                      className="flex w-28 flex-col items-center text-center sm:w-32"
                    >
                      <Avatar className={cn("size-16 ring-2 sm:size-20", ring)}>
                        {dev.avatar_url && (
                          <AvatarImage src={dev.avatar_url} alt={dev.github_username} />
                        )}
                        <AvatarFallback>{dev.github_username[0]?.toUpperCase()}</AvatarFallback>
                      </Avatar>
                      <p className="mt-2 truncate text-sm font-semibold">
                        {dev.display_name ?? dev.github_username}
                      </p>
                      <p className="text-xs text-muted-foreground">{label}</p>
                      <p className="mt-1 text-sm font-bold tabular-nums">
                        {sortValue(dev, sortBy).toLocaleString()}
                      </p>
                      <div
                        className={cn(
                          "mt-3 w-full rounded-t-lg bg-primary/15",
                          bar
                        )}
                      />
                    </Link>
                  );
                })}
              </div>
            </div>
          )}

          <div className="overflow-hidden rounded-2xl border bg-background">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/40 text-left text-muted-foreground">
                  <th className="px-4 py-3 font-medium">Rank</th>
                  <th className="px-4 py-3 font-medium">Developer</th>
                  <th className="hidden px-4 py-3 font-medium sm:table-cell">Score</th>
                  <th className="hidden px-4 py-3 font-medium md:table-cell">Stars</th>
                  <th className="hidden px-4 py-3 font-medium lg:table-cell">Repos</th>
                </tr>
              </thead>
              <tbody>
                {(top3.length < 3 ? developers : rest).map((dev, i) => {
                  const rank = top3.length < 3 ? i + 1 : i + 4;
                  return (
                    <tr key={dev.id} className="border-b last:border-0 hover:bg-muted/30">
                      <td className="px-4 py-3 font-mono text-muted-foreground">#{rank}</td>
                      <td className="px-4 py-3">
                        <Link
                          href={`/developers/${dev.github_username}`}
                          className="flex items-center gap-3 hover:underline"
                        >
                          <Avatar className="size-8">
                            {dev.avatar_url && (
                              <AvatarImage src={dev.avatar_url} alt={dev.github_username} />
                            )}
                            <AvatarFallback>
                              {dev.github_username[0]?.toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div>
                            <p className="font-medium">
                              {dev.display_name ?? dev.github_username}
                            </p>
                            <p className="text-xs text-muted-foreground">
                              @{dev.github_username}
                            </p>
                          </div>
                        </Link>
                      </td>
                      <td className="hidden px-4 py-3 tabular-nums sm:table-cell">
                        {Math.round(dev.activity_score)}
                      </td>
                      <td className="hidden px-4 py-3 tabular-nums md:table-cell">
                        {dev.total_stars}
                      </td>
                      <td className="hidden px-4 py-3 tabular-nums lg:table-cell">
                        {dev.public_repos}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  );
}
