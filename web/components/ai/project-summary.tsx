"use client";

import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";
import type { PublicRepo, ProjectSummary } from "@/lib/api";
import { fetchProjectSummary } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function ProjectImpactSummaryCard({ repo }: { repo: PublicRepo }) {
  const [summary, setSummary] = useState<ProjectSummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetchProjectSummary(repo.id)
      .then((data) => {
        if (!cancelled) {
          setSummary(data);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [repo.id]);

  if (loading) {
    return (
      <div className="mt-3 space-y-2 rounded-xl border border-primary/10 bg-primary/5 p-3">
        <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <Sparkles className="size-3.5" />
          AI project impact
        </div>
        <Skeleton className="h-3 w-40" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-11/12" />
      </div>
    );
  }

  if (!summary) {
    return (
      <div className="mt-3 rounded-xl border border-dashed border-border/70 bg-muted/20 p-3">
        <p className="text-[11px] uppercase tracking-wider text-muted-foreground">
          AI project impact
        </p>
        <p className="mt-1 text-xs text-muted-foreground">Summary unavailable for this repo.</p>
      </div>
    );
  }

  return (
    <div className="mt-3 rounded-xl border border-primary/10 bg-primary/5 p-3">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Sparkles className="size-3.5" />
        AI project impact
      </div>
      <p className="mt-2 text-sm font-medium leading-snug">{summary.headline}</p>
      <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{summary.summary}</p>
    </div>
  );
}
