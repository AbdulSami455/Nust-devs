"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import {
  api,
  type Developer,
  type DeveloperComparison,
} from "@/lib/api";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

type Metric = {
  label: string;
  left: number;
  right: number;
  format?: (value: number) => string;
};

function formatNumber(value: number) {
  return value.toLocaleString();
}

function compareLabel(metric: Metric) {
  if (metric.left === metric.right) return "tie";
  return metric.left > metric.right ? "left" : "right";
}

function DeveloperCard({
  dev,
  rank,
  label,
}: {
  dev: Developer;
  rank?: number;
  label: string;
}) {
  return (
    <Card className="h-full">
      <CardHeader className="pb-3">
        <div className="flex items-start gap-3">
          <Avatar className="size-14">
            {dev.avatar_url && <AvatarImage src={dev.avatar_url} alt={dev.github_username} />}
            <AvatarFallback>{dev.github_username[0]?.toUpperCase()}</AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <CardTitle className="truncate">
                {dev.display_name ?? dev.github_username}
              </CardTitle>
              {rank != null && <Badge variant="secondary">#{rank}</Badge>}
            </div>
            <CardDescription className="font-mono">@{dev.github_username}</CardDescription>
            <p className="mt-1 text-sm text-muted-foreground">{label}</p>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-3">
          {[
            ["Stars", dev.total_stars],
            ["Repos", dev.public_repos],
            ["Followers", dev.followers],
            ["Activity", Math.round(dev.activity_score)],
          ].map(([name, value]) => (
            <div key={name} className="rounded-xl border bg-muted/30 px-3 py-2">
              <p className="text-xs text-muted-foreground">{name}</p>
              <p className="text-lg font-semibold tabular-nums">{Number(value).toLocaleString()}</p>
            </div>
          ))}
        </div>

        <div className="flex flex-wrap gap-2">
          <Link
            href={`/developers/${dev.github_username}`}
            className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
          >
            View profile
          </Link>
          <a
            href={`https://github.com/${dev.github_username}`}
            target="_blank"
            rel="noopener noreferrer"
            className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
          >
            GitHub
          </a>
        </div>
      </CardContent>
    </Card>
  );
}

export function CompareClient() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [developers, setDevelopers] = useState<Developer[]>([]);
  const [left, setLeft] = useState("");
  const [right, setRight] = useState("");
  const [comparison, setComparison] = useState<DeveloperComparison | null>(null);
  const [loadingDevList, setLoadingDevList] = useState(true);
  const [loadingComparison, setLoadingComparison] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const leftParam = searchParams.get("left") ?? "";
  const rightParam = searchParams.get("right") ?? "";

  useEffect(() => {
    api.public.developers
      .list(1, 100)
      .then(setDevelopers)
      .catch(() => setDevelopers([]))
      .finally(() => setLoadingDevList(false));
  }, []);

  useEffect(() => {
    if (!developers.length) return;

    if (leftParam) {
      setLeft(leftParam);
    }
    if (rightParam) {
      setRight(rightParam);
    } else if (leftParam) {
      const fallback = developers.find((dev) => dev.github_username !== leftParam)?.github_username;
      setRight(fallback ?? "");
    }
  }, [developers, leftParam, rightParam]);

  useEffect(() => {
    if (!left || !right) {
      setComparison(null);
      setError(null);
      return;
    }
    if (left === right) {
      setComparison(null);
      setError("Pick two different developers.");
      return;
    }

    let cancelled = false;
    setLoadingComparison(true);
    setError(null);

    api.ai
      .compareDevelopers(left, right)
      .then((data) => {
        if (!cancelled) {
          setComparison(data);
        }
      })
      .catch((err: Error) => {
        if (!cancelled) {
          setComparison(null);
          setError(err.message || "Failed to compare developers");
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoadingComparison(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [left, right]);

  const leftDev = useMemo(
    () => developers.find((dev) => dev.github_username === left) ?? comparison?.left ?? null,
    [comparison?.left, developers, left]
  );
  const rightDev = useMemo(
    () => developers.find((dev) => dev.github_username === right) ?? comparison?.right ?? null,
    [comparison?.right, developers, right]
  );

  const metrics = useMemo<Metric[]>(
    () =>
      leftDev && rightDev
        ? [
            { label: "Activity score", left: leftDev.activity_score, right: rightDev.activity_score },
            { label: "Stars", left: leftDev.total_stars, right: rightDev.total_stars },
            { label: "Repos", left: leftDev.public_repos, right: rightDev.public_repos },
            { label: "Followers", left: leftDev.followers, right: rightDev.followers },
            { label: "Builder", left: leftDev.builder_score, right: rightDev.builder_score },
            { label: "Contributor", left: leftDev.contributor_score, right: rightDev.contributor_score },
            { label: "Reviewer", left: leftDev.reviewer_score, right: rightDev.reviewer_score },
            { label: "Community", left: leftDev.community_score, right: rightDev.community_score },
            { label: "PRs", left: leftDev.pr_contributions, right: rightDev.pr_contributions },
            { label: "Reviews", left: leftDev.review_contributions, right: rightDev.review_contributions },
            { label: "Current streak", left: leftDev.current_streak ?? 0, right: rightDev.current_streak ?? 0 },
          ]
        : [],
    [leftDev, rightDev]
  );

  const loading = loadingDevList || loadingComparison;

  const submit = () => {
    if (!left || !right) return;
    router.push(`/compare?left=${encodeURIComponent(left)}&right=${encodeURIComponent(right)}`);
  };

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-4 py-8 sm:px-6">
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight">Compare developers</h1>
        <p className="max-w-2xl text-sm text-muted-foreground">
          Pick two tracked developers and get a side-by-side breakdown with AI-written takeaways.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Choose developers</CardTitle>
          <CardDescription>
            The comparison uses public profile data, repo activity, and contribution history.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form
            className="grid gap-3 md:grid-cols-[1fr_1fr_auto]"
            onSubmit={(e) => {
              e.preventDefault();
              submit();
            }}
          >
            <label className="space-y-2">
              <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                Left developer
              </span>
              <select
                value={left}
                onChange={(e) => setLeft(e.target.value)}
                className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
              >
                <option value="">Select developer</option>
                {developers.map((dev) => (
                  <option key={dev.id} value={dev.github_username}>
                    {dev.display_name ?? dev.github_username} (@{dev.github_username})
                  </option>
                ))}
              </select>
            </label>

            <label className="space-y-2">
              <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                Right developer
              </span>
              <select
                value={right}
                onChange={(e) => setRight(e.target.value)}
                className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
              >
                <option value="">Select developer</option>
                {developers.map((dev) => (
                  <option key={dev.id} value={dev.github_username}>
                    {dev.display_name ?? dev.github_username} (@{dev.github_username})
                  </option>
                ))}
              </select>
            </label>

            <div className="flex items-end">
              <Button type="submit" className="w-full md:w-auto" disabled={!left || !right || left === right}>
                Compare
              </Button>
            </div>
          </form>
          {error && <p className="mt-3 text-sm text-rose-600 dark:text-rose-400">{error}</p>}
        </CardContent>
      </Card>

      {!comparison && loading ? (
        <div className="space-y-4">
          <Skeleton className="h-36 rounded-2xl" />
          <Skeleton className="h-60 rounded-2xl" />
        </div>
      ) : comparison && leftDev && rightDev ? (
        <div className="space-y-6">
          <Card className="border-primary/20 bg-gradient-to-br from-primary/5 via-card to-card">
            <CardHeader>
              <div className="flex flex-wrap items-center gap-2">
                <CardTitle>{comparison.headline}</CardTitle>
                <Badge variant={comparison.source === "ai" ? "default" : "secondary"}>
                  {comparison.source === "ai" ? "AI takeaways" : "Fallback takeaways"}
                </Badge>
              </div>
              <CardDescription>{comparison.summary}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {comparison.takeaways.length > 0 && (
                <div className="space-y-2">
                  <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                    Key takeaways
                  </p>
                  <ul className="space-y-2">
                    {comparison.takeaways.map((item) => (
                      <li key={item} className="rounded-xl border bg-background/70 px-3 py-2 text-sm">
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              <div className="flex flex-wrap gap-2">
                {comparison.shared_strengths.map((item) => (
                  <Badge key={item} variant="outline">
                    Shared: {item}
                  </Badge>
                ))}
                <Badge variant="secondary">{comparison.verdict}</Badge>
              </div>
            </CardContent>
          </Card>

          <div className="grid gap-4 lg:grid-cols-2">
            <DeveloperCard
              dev={leftDev}
              rank={comparison.left_rank}
              label="Left side"
            />
            <DeveloperCard
              dev={rightDev}
              rank={comparison.right_rank}
              label="Right side"
            />
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Metric breakdown</CardTitle>
              <CardDescription>Highlighted values show which developer leads for each metric.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {metrics.map((metric) => {
                const winner = compareLabel(metric);
                return (
                  <div
                    key={metric.label}
                    className="grid grid-cols-[1fr_auto_1fr] items-center gap-3 rounded-xl border px-3 py-2"
                  >
                    <p
                      className={cn(
                        "font-medium tabular-nums",
                        winner === "left" ? "text-primary" : "text-foreground"
                      )}
                    >
                      {metric.format ? metric.format(metric.left) : formatNumber(metric.left)}
                    </p>
                    <p className="text-center text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                      {metric.label}
                    </p>
                    <p
                      className={cn(
                        "text-right font-medium tabular-nums",
                        winner === "right" ? "text-primary" : "text-foreground"
                      )}
                    >
                      {metric.format ? metric.format(metric.right) : formatNumber(metric.right)}
                    </p>
                  </div>
                );
              })}
            </CardContent>
          </Card>
        </div>
      ) : (
        <Card className="border-dashed">
          <CardContent className="py-16 text-center">
            <p className="text-sm text-muted-foreground">
              Pick two developers to generate a side-by-side comparison.
            </p>
            <div className="mt-4 flex flex-wrap justify-center gap-2">
              <Link href="/developers" className={cn(buttonVariants({ variant: "outline" }))}>
                Browse developers
              </Link>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
