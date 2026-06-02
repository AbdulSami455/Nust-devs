"use client";

import type { Overview } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import { GitBranch, Star, Users, Zap } from "lucide-react";

const statConfig = [
  { key: "total_developers" as const, label: "Developers", icon: Users },
  { key: "total_repos" as const, label: "Repositories", icon: GitBranch },
  { key: "total_stars" as const, label: "Total Stars", icon: Star },
  { key: "total_contributions" as const, label: "Contributions", icon: Zap },
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
      {statConfig.map(({ key, label, icon: Icon }, i) => (
        <div
          key={key}
          className={`bento-card flex flex-col gap-3 ${i === 0 ? "lg:col-span-1" : ""}`}
        >
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              {label}
            </span>
            <Icon className="size-4 text-primary/80" />
          </div>
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
