"use client";

import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";
import { fetchHomeBanner, type HomeBannerInsight } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function HomeBannerCard() {
  const [banner, setBanner] = useState<HomeBannerInsight | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetchHomeBanner()
      .then((data) => {
        if (!cancelled) setBanner(data);
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
          AI what changed today
        </div>
        <Skeleton className="h-5 w-56" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-11/12" />
        <div className="flex flex-wrap gap-2">
          <Skeleton className="h-5 w-28 rounded-full" />
          <Skeleton className="h-5 w-24 rounded-full" />
          <Skeleton className="h-5 w-32 rounded-full" />
        </div>
      </section>
    );
  }

  if (!banner) {
    return null;
  }

  return (
    <section className="bento-card space-y-4 border-primary/15 bg-gradient-to-br from-primary/5 via-card to-card">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Sparkles className="size-3.5" />
        AI what changed today
      </div>
      <div>
        <p className="text-base font-medium leading-snug">{banner.headline}</p>
        <p className="mt-1 text-sm leading-relaxed text-muted-foreground">{banner.summary}</p>
      </div>
      {banner.highlights.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {banner.highlights.map((item) => (
            <span
              key={item}
              className="inline-flex items-center rounded-full border border-border bg-background/70 px-2.5 py-0.5 text-xs text-foreground"
            >
              {item}
            </span>
          ))}
        </div>
      )}
    </section>
  );
}
