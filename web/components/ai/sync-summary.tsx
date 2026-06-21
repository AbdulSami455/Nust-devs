"use client";

import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";
import { fetchAdminSyncSummary, type SyncSummary } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function SyncSummaryCard() {
  const [summary, setSummary] = useState<SyncSummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetchAdminSyncSummary()
      .then((data) => {
        if (!cancelled) setSummary(data);
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
      <section className="space-y-3 rounded-xl border border-primary/10 bg-primary/5 p-3">
        <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <Sparkles className="size-3.5" />
          Sync summary
        </div>
        <Skeleton className="h-4 w-44" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-11/12" />
      </section>
    );
  }

  if (!summary) return null;

  return (
    <section className="space-y-3 rounded-xl border border-primary/10 bg-primary/5 p-3">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Sparkles className="size-3.5" />
        Sync summary
      </div>
      <p className="text-sm font-medium leading-snug">{summary.headline}</p>
      <p className="text-xs leading-relaxed text-muted-foreground">{summary.summary}</p>
      <div className="flex flex-wrap gap-2">
        {summary.highlights.map((item) => (
          <span key={item} className="rounded-full border border-border bg-background/70 px-2.5 py-0.5 text-xs">
            {item}
          </span>
        ))}
      </div>
    </section>
  );
}
