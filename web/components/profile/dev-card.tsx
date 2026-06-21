"use client";

import { useCallback } from "react";
import Link from "next/link";
import type { Developer } from "@/lib/api";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import { PowerLevelBadge } from "@/components/profile/power-level";

export function DevCard({
  dev,
  rank,
}: {
  dev: Developer;
  rank?: number;
}) {
  const profileUrl =
    typeof window !== "undefined"
      ? `${window.location.origin}/developers/${dev.github_username}`
      : `/developers/${dev.github_username}`;

  const copyLink = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(profileUrl);
      toast.success("Profile link copied");
    } catch {
      toast.error("Could not copy link");
    }
  }, [profileUrl]);

  const shareTwitter = useCallback(() => {
    const text = `${dev.display_name ?? dev.github_username} — NUST developer on GitHub. ${dev.total_stars} stars, ${dev.public_repos} repos.`;
    const url = `https://twitter.com/intent/tweet?text=${encodeURIComponent(text)}&url=${encodeURIComponent(profileUrl)}`;
    window.open(url, "_blank", "noopener,noreferrer");
  }, [dev, profileUrl]);

  return (
    <div
      id="dev-card"
      className="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-card via-card to-primary/5 p-6 sm:p-8"
    >
      <div className="absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-primary via-[var(--nust-gold)] to-primary" />

      <div className="flex flex-col gap-6 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex gap-4">
          <Avatar className="size-20 border-2 border-background shadow-md">
            {dev.avatar_url && <AvatarImage src={dev.avatar_url} alt={dev.github_username} />}
            <AvatarFallback className="text-2xl">
              {dev.github_username[0]?.toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0 space-y-1">
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
                {dev.display_name ?? dev.github_username}
              </h1>
              {rank != null && rank <= 10 && (
                <Badge variant="secondary">Rank #{rank}</Badge>
              )}
              {dev.verification_status === "email_verified" && (
                <Badge>Verified</Badge>
              )}
            </div>
            <p className="font-mono text-sm text-muted-foreground">@{dev.github_username}</p>
            <div className="pt-2">
              <PowerLevelBadge dev={dev} />
            </div>
            {dev.bio && (
              <p className="max-w-lg pt-1 text-sm text-muted-foreground">{dev.bio}</p>
            )}
            <div className="flex flex-wrap gap-x-4 gap-y-1 pt-2 text-sm text-muted-foreground">
              {dev.location && <span>{dev.location}</span>}
              {dev.company && <span>{dev.company}</span>}
              {dev.website && (
                <a
                  href={dev.website.startsWith("http") ? dev.website : `https://${dev.website}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  {dev.website.replace(/^https?:\/\//, "")}
                </a>
              )}
            </div>
          </div>
        </div>

        <div className="flex shrink-0 flex-wrap gap-2">
          <Button variant="outline" onClick={copyLink}>
            Copy link
          </Button>
          <Button variant="outline" onClick={shareTwitter}>
            Share on X
          </Button>
          <Link
            href={`/compare?left=${dev.github_username}`}
            className={cn(buttonVariants({ variant: "outline" }))}
          >
            Compare
          </Link>
          <Link
            href={`/developers/${dev.github_username}/wrapped`}
            className={cn(buttonVariants({ variant: "outline" }))}
          >
            {new Date().getFullYear()} Wrapped
          </Link>
          <a
            href={`https://github.com/${dev.github_username}`}
            target="_blank"
            rel="noopener noreferrer"
            className={cn(buttonVariants())}
          >
            GitHub profile
          </a>
        </div>
      </div>

      <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[
          { label: "Stars", value: dev.total_stars },
          { label: "Repositories", value: dev.public_repos },
          { label: "Followers", value: dev.followers },
          { label: "Activity score", value: Math.round(dev.activity_score) },
        ].map(({ label, value }) => (
          <div key={label} className="rounded-xl border bg-background/60 px-4 py-3">
            <p className="text-xs text-muted-foreground">{label}</p>
            <p className="text-xl font-bold tabular-nums">{value.toLocaleString()}</p>
          </div>
        ))}
      </div>

      <div className="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[
          { label: "Builder", value: Math.round(dev.builder_score ?? 0) },
          { label: "Contributor", value: Math.round(dev.contributor_score ?? 0) },
          { label: "Reviewer", value: Math.round(dev.reviewer_score ?? 0) },
          { label: "Community", value: Math.round(dev.community_score ?? 0) },
        ].map(({ label, value }) => (
          <div key={label} className="rounded-xl border border-dashed bg-muted/30 px-4 py-3">
            <p className="text-xs text-muted-foreground">{label}</p>
            <p className="text-lg font-semibold tabular-nums">{value.toLocaleString()}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
