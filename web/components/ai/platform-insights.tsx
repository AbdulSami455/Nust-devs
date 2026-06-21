"use client";

import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";
import { fetchPlatformInsights, type PlatformInsights } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function PlatformInsightsCard() {
  const [insight, setInsight] = useState<PlatformInsights | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetchPlatformInsights()
      .then((data) => {
        if (!cancelled) setInsight(data);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  if (loading) {
    return (
      <section className="bento-card space-y-3 border-primary/15 bg-gradient-to-br from-primary/5 via-card to-card">
        <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <Sparkles className="size-3.5" />
          AI platform insights
        </div>
        <Skeleton className="h-4 w-56" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-11/12" />
      </section>
    );
  }

  if (!insight) return null;

  return (
    <section className="bento-card space-y-4 border-primary/15 bg-gradient-to-br from-primary/5 via-card to-card">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Sparkles className="size-3.5" />
        AI platform insights
      </div>
      <p className="text-base font-medium leading-snug">{insight.headline}</p>
      <div className="space-y-3">
        {insight.project_insights.length > 0 && (
          <Block label="Projects" items={insight.project_insights} />
        )}
        {insight.community_trends.length > 0 && (
          <Block label="Community trends" items={insight.community_trends} />
        )}
      </div>
    </section>
  );
}

function Block({ label, items }: { label: string; items: string[] }) {
  return (
    <div className="space-y-2">
      <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">{label}</p>
      <div className="flex flex-wrap gap-2">
        {items.map((item) => (
          <span key={item} className="rounded-full border border-border bg-background/70 px-2.5 py-0.5 text-xs">
            {item}
          </span>
        ))}
      </div>
    </div>
  );
}
