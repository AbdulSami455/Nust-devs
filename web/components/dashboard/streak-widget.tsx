"use client";

import Link from "next/link";
import type { StreakSummary } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function StreakWidget({
  summary,
  loading,
}: {
  summary: StreakSummary | null;
  loading: boolean;
}) {
  if (loading) {
    return <Skeleton className="h-20 rounded-2xl" />;
  }
  if (!summary || summary.devs_on_7plus_streak === 0) {
    return null;
  }

  return (
    <Link
      href="/leaderboard?sort=streak"
      className="bento-card flex items-center justify-between gap-4 transition-colors hover:border-primary/40"
    >
      <div>
        <p className="text-sm font-medium">
          {summary.devs_on_7plus_streak} dev
          {summary.devs_on_7plus_streak === 1 ? "" : "s"} on a 7+ day streak
        </p>
        <p className="text-xs text-muted-foreground">
          Longest active streak: {summary.longest_active_streak} days
          {summary.devs_on_30plus_streak > 0 &&
            ` · ${summary.devs_on_30plus_streak} at 30+ days`}
        </p>
      </div>
      <span className="text-2xl" aria-hidden>
        🔥
      </span>
    </Link>
  );
}
