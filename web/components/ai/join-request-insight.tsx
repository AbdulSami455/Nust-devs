"use client";

import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";
import { fetchJoinRequestInsight, type JoinRequestInsight } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function JoinRequestInsightCard({ requestId }: { requestId: string }) {
  const [insight, setInsight] = useState<JoinRequestInsight | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetchJoinRequestInsight(requestId)
      .then((data) => {
        if (!cancelled) setInsight(data);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [requestId]);

  if (loading) {
    return (
      <section className="space-y-3 rounded-xl border border-primary/10 bg-primary/5 p-3">
        <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <Sparkles className="size-3.5" />
          AI request insight
        </div>
        <Skeleton className="h-4 w-48" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-11/12" />
      </section>
    );
  }

  if (!insight) return null;

  return (
    <section className="space-y-3 rounded-xl border border-primary/10 bg-primary/5 p-3">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Sparkles className="size-3.5" />
        AI request insight
      </div>
      <p className="text-sm font-medium leading-snug">{insight.headline}</p>
      <p className="text-xs leading-relaxed text-muted-foreground">{insight.summary}</p>
      <p className="text-xs font-medium text-amber-700 dark:text-amber-400">{insight.duplicate_warning}</p>
      {insight.matched_username && (
        <p className="text-[11px] uppercase tracking-wider text-muted-foreground">
          Closest profile: @{insight.matched_username}
        </p>
      )}
    </section>
  );
}
