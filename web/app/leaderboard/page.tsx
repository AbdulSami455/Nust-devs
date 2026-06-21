"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, type LeaderboardEntry } from "@/lib/api";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Skeleton } from "@/components/ui/skeleton";
import { RankDelta, Sparkline } from "@/components/charts/sparkline";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const SORT_OPTIONS = [
  { value: "activity_score", label: "Activity" },
  { value: "streak", label: "Streak" },
  { value: "power_level", label: "Level" },
  { value: "xp", label: "XP" },
  { value: "builder_score", label: "Builder" },
  { value: "contributor_score", label: "Contributor" },
  { value: "reviewer_score", label: "Reviewer" },
  { value: "community_score", label: "Community" },
  { value: "total_stars", label: "Stars" },
  { value: "public_repos", label: "Repos" },
  { value: "followers", label: "Followers" },
];

const LIMIT_TABS = [
  { value: 10, label: "Top 10" },
  { value: 50, label: "Top 50" },
  { value: 100, label: "All" },
];

const VIEW_TABS = [
  { value: "default" as const, label: "Rankings" },
  { value: "rising" as const, label: "Rising (7d)", period: 7 as const },
  { value: "rising30" as const, label: "Rising (30d)", period: 30 as const },
];

const PODIUM = [
  { place: 2, ring: "ring-slate-400/60", bar: "h-20", label: "2nd" },
  { place: 1, ring: "ring-[var(--nust-gold)]/80", bar: "h-28", label: "1st" },
  { place: 3, ring: "ring-amber-700/60", bar: "h-16", label: "3rd" },
];

function sortValue(dev: LeaderboardEntry, sortBy: string) {
  switch (sortBy) {
    case "total_stars":
      return dev.total_stars;
    case "public_repos":
      return dev.public_repos;
    case "followers":
      return dev.followers;
    case "builder_score":
      return Math.round(dev.builder_score);
    case "contributor_score":
      return Math.round(dev.contributor_score);
    case "reviewer_score":
      return Math.round(dev.reviewer_score);
    case "community_score":
      return Math.round(dev.community_score);
    case "streak":
      return dev.current_streak ?? 0;
    case "power_level":
      return dev.power_level ?? 1;
    case "xp":
      return dev.xp ?? 0;
    default:
      return Math.round(dev.activity_score);
  }
}

function sortLabel(sortBy: string) {
  return SORT_OPTIONS.find((o) => o.value === sortBy)?.label ?? "Score";
}

function scoreDelta(dev: LeaderboardEntry, period: 7 | 30) {
  const d = period === 30 ? dev.score_delta_30d : dev.score_delta_7d;
  return d ?? null;
}

function rankDelta(dev: LeaderboardEntry, period: 7 | 30) {
  const d = period === 30 ? dev.rank_delta_30d : dev.rank_delta_7d;
  return d ?? null;
}

export default function LeaderboardPage() {
  const [developers, setDevelopers] = useState<LeaderboardEntry[]>([]);
  const [sortBy, setSortBy] = useState("activity_score");
  const [limit, setLimit] = useState(50);
  const [view, setView] = useState<"default" | "rising" | "rising30">("default");
  const [loading, setLoading] = useState(true);

  const trendPeriod: 7 | 30 = view === "rising30" ? 30 : 7;
  const isRising = view !== "default";

  useEffect(() => {
    setLoading(true);
    api.public
      .leaderboard(sortBy, 1, limit, {
        view: isRising ? "rising" : "default",
        period: trendPeriod,
      })
      .then(setDevelopers)
      .catch(() => setDevelopers([]))
      .finally(() => setLoading(false));
  }, [sortBy, limit, view, isRising, trendPeriod]);

  const showPodium = !isRising;
  const top3 = showPodium ? developers.slice(0, 3) : [];
  const tableRows = showPodium && top3.length >= 3 ? developers.slice(3) : developers;

  return (
    <div className="mx-auto max-w-5xl space-y-8 px-4 py-8 sm:px-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">Leaderboard</h1>
        <p className="text-sm text-muted-foreground">
          Rankings among tracked NUST developers — with rank trends and score momentum.
        </p>
      </div>

      <div className="flex flex-wrap gap-2">
        {VIEW_TABS.map((tab) => (
          <button
            key={tab.value}
            type="button"
            onClick={() => setView(tab.value)}
            className={cn(
              "rounded-lg px-3 py-1.5 text-sm font-medium transition-colors",
              view === tab.value
                ? "bg-primary text-primary-foreground"
                : "border bg-background hover:bg-muted"
            )}
          >
            {tab.label}
          </button>
        ))}
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

      {isRising && (
        <p className="text-sm text-muted-foreground">
          Developers with the biggest score gains in the last {trendPeriod} days (from daily snapshots).
        </p>
      )}

      {loading ? (
        <Skeleton className="h-64 rounded-2xl" />
      ) : developers.length === 0 ? (
        <div className="rounded-2xl border border-dashed px-6 py-16 text-center text-muted-foreground">
          {isRising
            ? "Not enough snapshot history yet — trends appear after a few days of syncs."
            : "No developers synced yet."}
        </div>
      ) : (
        <>
          {showPodium && top3.length >= 3 && (
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
                      <div className="mt-1">
                        <RankDelta delta={rankDelta(dev, 7)} period={7} />
                      </div>
                      <div
                        className={cn("mt-3 w-full rounded-t-lg bg-primary/15", bar)}
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
                  <th className="hidden px-4 py-3 font-medium sm:table-cell">Trend</th>
                  {sortBy === "streak" && (
                    <th className="hidden px-4 py-3 font-medium md:table-cell">Multiplier</th>
                  )}
                  {isRising ? (
                    <th className="px-4 py-3 font-medium">Gain</th>
                  ) : (
                    <th className="hidden px-4 py-3 font-medium md:table-cell">Δ Rank</th>
                  )}
                  <th className="hidden px-4 py-3 font-medium sm:table-cell">{sortLabel(sortBy)}</th>
                  <th className="hidden px-4 py-3 font-medium md:table-cell">Stars</th>
                  <th className="hidden px-4 py-3 font-medium lg:table-cell">Repos</th>
                </tr>
              </thead>
              <tbody>
                {tableRows.map((dev, i) => {
                  const rank = dev.rank ?? (showPodium && top3.length >= 3 ? i + 4 : i + 1);
                  const delta = scoreDelta(dev, trendPeriod);
                  return (
                    <tr key={dev.id} className="border-b last:border-0 hover:bg-muted/30">
                      <td className="px-4 py-3 font-mono text-muted-foreground">#{rank}</td>
                      <td className="px-4 py-3">
                        <div className="flex items-start gap-3">
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
                          <Link
                            href={`/compare?left=${dev.github_username}`}
                            className={cn(
                              buttonVariants({ variant: "ghost", size: "xs" }),
                              "mt-0.5 px-0 text-xs text-primary hover:text-primary"
                            )}
                          >
                            Compare
                          </Link>
                        </div>
                      </td>
                      <td className="hidden px-4 py-3 sm:table-cell">
                        <Sparkline
                          data={dev.sparkline ?? []}
                          positive={
                            delta != null ? delta >= 0 : undefined
                          }
                        />
                      </td>
                      {sortBy === "streak" && (
                        <td className="hidden px-4 py-3 tabular-nums md:table-cell">
                          {(dev.streak_multiplier ?? 1) > 1
                            ? `${dev.streak_multiplier}x`
                            : "—"}
                        </td>
                      )}
                      {isRising ? (
                        <td className="px-4 py-3">
                          {delta != null ? (
                            <span
                              className={cn(
                                "font-medium tabular-nums",
                                delta > 0
                                  ? "text-emerald-600 dark:text-emerald-400"
                                  : delta < 0
                                    ? "text-rose-600 dark:text-rose-400"
                                    : "text-muted-foreground"
                              )}
                            >
                              {delta > 0 ? "+" : ""}
                              {Number.isInteger(delta) ? delta : delta.toFixed(1)}
                            </span>
                          ) : (
                            <span className="text-muted-foreground">—</span>
                          )}
                        </td>
                      ) : (
                        <td className="hidden px-4 py-3 md:table-cell">
                          <RankDelta delta={rankDelta(dev, 7)} period={7} />
                        </td>
                      )}
                      <td className="hidden px-4 py-3 tabular-nums sm:table-cell">
                        {sortValue(dev, sortBy).toLocaleString()}
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
