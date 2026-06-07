"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, type WrappedReport } from "@/lib/api";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Button, buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

export function WrappedClient({
  username,
  year: yearProp,
}: {
  username: string;
  year?: number;
}) {
  const [report, setReport] = useState<WrappedReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    api.public.developers
      .wrapped(username, yearProp)
      .then(setReport)
      .catch(() => setNotFound(true))
      .finally(() => setLoading(false));
  }, [username, yearProp]);

  if (loading) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-12">
        <Skeleton className="h-96 rounded-3xl" />
      </div>
    );
  }

  if (notFound || !report) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-24 text-center">
        <p className="font-semibold">Wrapped not available</p>
        <Link href={`/developers/${username}`} className="mt-4 inline-block text-primary hover:underline">
          Back to profile
        </Link>
      </div>
    );
  }

  const share = async () => {
    const url = window.location.href;
    try {
      await navigator.clipboard.writeText(url);
      toast.success("Wrapped link copied");
    } catch {
      toast.error("Could not copy link");
    }
  };

  return (
    <div className="mx-auto max-w-2xl px-4 py-8 sm:py-12">
      <div className="overflow-hidden rounded-3xl border bg-gradient-to-br from-primary/20 via-background to-[var(--nust-gold)]/15 p-6 sm:p-10">
        <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
          NUST Devs Wrapped · {report.year}
        </p>
        <div className="mt-6 flex items-center gap-4">
          <Avatar className="size-16 ring-2 ring-primary/30">
            {report.avatar_url && (
              <AvatarImage src={report.avatar_url} alt={report.username} />
            )}
            <AvatarFallback>{report.username[0]?.toUpperCase()}</AvatarFallback>
          </Avatar>
          <div>
            <h1 className="text-2xl font-bold">
              {report.display_name ?? report.username}
            </h1>
            <p className="text-sm text-muted-foreground">@{report.username}</p>
          </div>
        </div>

        <div className="mt-8 grid grid-cols-2 gap-3">
          <Stat label="Contributions" value={report.total_contributions} />
          <Stat label="Stars" value={report.total_stars} />
          <Stat label="Level" value={`${report.power_level} ${report.power_title}`} />
          <Stat label="Percentile" value={`Top ${100 - report.activity_percentile}%`} />
        </div>

        {report.top_repo && (
          <div className="mt-6 rounded-2xl border bg-background/60 p-4">
            <p className="text-xs text-muted-foreground">Top repository</p>
            <p className="font-mono text-sm font-medium">{report.top_repo}</p>
            <p className="text-xs text-muted-foreground">{report.top_repo_stars} stars</p>
          </div>
        )}

        {report.rank_change !== 0 && report.rank_start > 0 && (
          <p className="mt-4 text-sm">
            Leaderboard: #{report.rank_start} → #{report.rank_end}
            {report.rank_change > 0 && (
              <span className="text-emerald-600 dark:text-emerald-400">
                {" "}
                (↑{report.rank_change})
              </span>
            )}
          </p>
        )}

        <ul className="mt-6 space-y-2">
          {report.highlights.map((h) => (
            <li key={h} className="flex items-start gap-2 text-sm">
              <span className="text-primary">•</span>
              {h}
            </li>
          ))}
        </ul>

        {report.top_languages.length > 0 && (
          <div className="mt-6 flex flex-wrap gap-2">
            {report.top_languages.map((l) => (
              <Badge key={l.name} variant="outline">
                {l.name}
              </Badge>
            ))}
          </div>
        )}

        <div className="mt-8 flex flex-wrap gap-2">
          <Button variant="outline" onClick={share}>
            Share wrapped
          </Button>
          <Link
            href={`/developers/${username}`}
            className={cn(buttonVariants({ variant: "secondary" }))}
          >
            View profile
          </Link>
        </div>
      </div>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="rounded-xl border bg-background/50 px-4 py-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="text-lg font-bold tabular-nums">{value}</p>
    </div>
  );
}
