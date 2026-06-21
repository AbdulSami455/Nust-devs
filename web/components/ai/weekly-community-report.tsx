"use client";

import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";
import { fetchWeeklyCommunityReport, type WeeklyCommunityReport } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function WeeklyCommunityReportCard() {
  const [report, setReport] = useState<WeeklyCommunityReport | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetchWeeklyCommunityReport()
      .then((data) => {
        if (!cancelled) setReport(data);
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
          Weekly community report
        </div>
        <Skeleton className="h-4 w-48" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-11/12" />
      </section>
    );
  }

  if (!report) return null;

  return (
    <section className="bento-card space-y-4 border-primary/15 bg-gradient-to-br from-primary/5 via-card to-card">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Sparkles className="size-3.5" />
        Weekly community report
      </div>
      <p className="text-base font-medium leading-snug">{report.headline}</p>
      <p className="text-sm leading-relaxed text-muted-foreground">{report.summary}</p>
      <div className="flex flex-wrap gap-2">
        {report.highlights.map((item) => (
          <span key={item} className="rounded-full border border-border bg-background/70 px-2.5 py-0.5 text-xs">
            {item}
          </span>
        ))}
      </div>
    </section>
  );
}
