"use client";

import type { Overview } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

const statConfig = [
  { key: "total_developers" as const, label: "Developers" },
  { key: "total_repos" as const, label: "Repositories" },
  { key: "total_stars" as const, label: "Total Stars" },
  { key: "total_contributions" as const, label: "Contributions" },
];

export function BentoStats({
  overview,
  loading,
}: {
  overview: Overview | null;
  loading: boolean;
}) {
  return (
    <div className="grid grid-cols-2 gap-3 lg:grid-cols-4 lg:gap-4">
      {statConfig.map(({ key, label }) => (
        <div key={key} className="bento-card flex flex-col gap-2">
          <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            {label}
          </span>
          {loading ? (
            <Skeleton className="h-9 w-24" />
          ) : (
            <p className="text-3xl font-bold tabular-nums tracking-tight">
              {(overview?.[key] ?? 0).toLocaleString()}
            </p>
          )}
        </div>
      ))}
    </div>
  );
}
