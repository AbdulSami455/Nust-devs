"use client";

import Link from "next/link";
import type { OSSStats } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const items = [
  { key: "original_projects" as const, label: "Original projects" },
  { key: "fork_projects" as const, label: "Forks tracked" },
  { key: "total_stars" as const, label: "Stars on originals" },
  { key: "contributors" as const, label: "Contributors" },
];

export function OSSStatsPanel({
  stats,
  loading,
}: {
  stats: OSSStats | null;
  loading: boolean;
}) {
  return (
    <div className="bento-card">
      <div className="mb-4 flex items-start justify-between gap-4">
        <div>
          <h2 className="text-base font-semibold tracking-tight">
            Open Source
          </h2>
          <p className="mt-1 max-w-2xl text-sm leading-6 text-muted-foreground">
            Original repos, forks, and community impact from NUST developers.
          </p>
        </div>
        <Link href="/projects" className={cn(buttonVariants({ variant: "outline" }))}>
          Browse projects
        </Link>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {items.map(({ key, label }) => (
          <div key={key} className="rounded-lg border bg-secondary/35 px-4 py-3">
            <p className="text-xs font-medium text-muted-foreground">{label}</p>
            {loading ? (
              <Skeleton className="mt-2 h-7 w-12" />
            ) : (
              <p className="mt-1 text-2xl font-semibold tabular-nums">
                {(stats?.[key] ?? 0).toLocaleString()}
              </p>
            )}
          </div>
        ))}
      </div>

      {!loading && stats?.top_language && (
        <p className="mt-4 text-sm text-muted-foreground">
          Most common language among original projects:{" "}
          <span className="font-medium text-foreground">{stats.top_language}</span>
        </p>
      )}
    </div>
  );
}
