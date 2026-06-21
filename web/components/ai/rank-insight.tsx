"use client";

import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";
import { fetchRankInsight, type RankInsight } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function RankInsightCard({ username }: { username: string }) {
  const [insight, setInsight] = useState<RankInsight | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetchRankInsight(username)
      .then((data) => {
        if (!cancelled) {
          setInsight(data);
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
  }, [username]);

  if (loading) {
    return (
      <div className="bento-card space-y-3">
        <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <Sparkles className="size-3.5" />
          AI rank & badge insight
        </div>
        <Skeleton className="h-4 w-52" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-11/12" />
        <div className="flex flex-wrap gap-2">
          <Skeleton className="h-5 w-24 rounded-full" />
          <Skeleton className="h-5 w-28 rounded-full" />
          <Skeleton className="h-5 w-24 rounded-full" />
        </div>
      </div>
    );
  }

  if (!insight) {
    return (
      <div className="bento-card space-y-2">
        <p className="text-[11px] uppercase tracking-wider text-muted-foreground">
          AI rank & badge insight
        </p>
        <p className="text-sm text-muted-foreground">Rank explanation unavailable right now.</p>
      </div>
    );
  }

  return (
    <section className="bento-card space-y-3 border-primary/15 bg-gradient-to-br from-primary/5 via-card to-card">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Sparkles className="size-3.5" />
        AI rank & badge insight
      </div>
      <p className="text-base font-medium leading-snug">{insight.headline}</p>
      <p className="text-sm leading-relaxed text-muted-foreground">{insight.summary}</p>
      {insight.highlights.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {insight.highlights.map((item) => (
            <span
              key={item}
              className="inline-flex items-center rounded-full border border-border bg-muted px-2.5 py-0.5 text-xs text-foreground"
            >
              {item}
            </span>
          ))}
        </div>
      )}
    </section>
  );
}
