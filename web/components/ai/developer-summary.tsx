"use client";

import { useEffect, useState } from "react";
import { fetchDeveloperSummary, type DeveloperSummary } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function DeveloperSummaryCard({ username }: { username: string }) {
  const [summary, setSummary] = useState<DeveloperSummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDeveloperSummary(username).then((s) => {
      setSummary(s);
      setLoading(false);
    });
  }, [username]);

  if (loading) {
    return (
      <div className="bento-card space-y-3">
        <div className="flex items-center gap-2">
          <Skeleton className="h-4 w-4 rounded-full" />
          <Skeleton className="h-3 w-24" />
        </div>
        <Skeleton className="h-4 w-3/4" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-5/6" />
        <div className="flex gap-2 flex-wrap">
          <Skeleton className="h-5 w-20 rounded-full" />
          <Skeleton className="h-5 w-28 rounded-full" />
          <Skeleton className="h-5 w-24 rounded-full" />
        </div>
      </div>
    );
  }

  if (!summary) return null;

  return (
    <div className="bento-card space-y-3">
      {/* Label */}
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <svg
          className="size-3.5 shrink-0"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
          />
        </svg>
        <span>AI-generated insight</span>
      </div>

      {/* Headline */}
      <p className="font-medium leading-snug">{summary.headline}</p>

      {/* Summary */}
      <p className="text-sm text-muted-foreground leading-relaxed">{summary.summary}</p>

      {/* Strengths */}
      {summary.strengths.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {summary.strengths.map((s) => (
            <span
              key={s}
              className="inline-flex items-center rounded-full border border-border bg-muted px-2.5 py-0.5 text-xs text-foreground"
            >
              {s}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
