"use client";

import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";
import { fetchProfileInsights, type ProfileInsights } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function ProfileInsightsCard({ username }: { username: string }) {
  const [insight, setInsight] = useState<ProfileInsights | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetchProfileInsights(username)
      .then((data) => {
        if (!cancelled) setInsight(data);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [username]);

  if (loading) {
    return (
      <section className="bento-card space-y-3">
        <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <Sparkles className="size-3.5" />
          AI profile insights
        </div>
        <Skeleton className="h-4 w-44" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-11/12" />
        <div className="flex flex-wrap gap-2">
          <Skeleton className="h-5 w-24 rounded-full" />
          <Skeleton className="h-5 w-28 rounded-full" />
          <Skeleton className="h-5 w-20 rounded-full" />
        </div>
      </section>
    );
  }

  if (!insight) {
    return (
      <section className="bento-card space-y-2">
        <p className="text-[11px] uppercase tracking-wider text-muted-foreground">
          AI profile insights
        </p>
        <p className="text-sm text-muted-foreground">Insights unavailable right now.</p>
      </section>
    );
  }

  return (
    <section className="bento-card space-y-4 border-primary/15 bg-gradient-to-br from-primary/5 via-card to-card">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Sparkles className="size-3.5" />
        AI profile insights
      </div>
      <div>
        <p className="text-base font-medium leading-snug">{insight.headline}</p>
        <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
          {insight.recent_activity_recap}
        </p>
      </div>

      {insight.top_achievements.length > 0 && (
        <Block label="Top achievements" items={insight.top_achievements} />
      )}
      {insight.completion_tips.length > 0 && (
        <Block label="Completion tips" items={insight.completion_tips} />
      )}
    </section>
  );
}

function Block({ label, items }: { label: string; items: string[] }) {
  return (
    <div className="space-y-2">
      <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
        {label}
      </p>
      <div className="flex flex-wrap gap-2">
        {items.map((item) => (
          <span
            key={item}
            className="inline-flex items-center rounded-full border border-border bg-background/70 px-2.5 py-0.5 text-xs text-foreground"
          >
            {item}
          </span>
        ))}
      </div>
    </div>
  );
}
