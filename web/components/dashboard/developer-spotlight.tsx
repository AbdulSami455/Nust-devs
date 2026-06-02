"use client";

import Link from "next/link";
import type { Developer } from "@/lib/api";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { ArrowRight, Sparkles } from "lucide-react";

export function DeveloperSpotlight({
  dev,
  loading,
}: {
  dev: Developer | null;
  loading: boolean;
}) {
  return (
    <div className="bento-card relative overflow-hidden">
      <div className="absolute -right-8 -top-8 size-32 rounded-full bg-primary/10 blur-2xl" />
      <div className="relative flex items-center gap-2 text-primary">
        <Sparkles className="size-4" />
        <span className="text-xs font-semibold uppercase tracking-wider">Developer Spotlight</span>
      </div>

      {loading ? (
        <div className="mt-4 space-y-3">
          <Skeleton className="size-16 rounded-full" />
          <Skeleton className="h-5 w-32" />
          <Skeleton className="h-4 w-48" />
        </div>
      ) : !dev ? (
        <p className="mt-4 text-sm text-muted-foreground">No synced developers yet.</p>
      ) : (
        <>
          <div className="mt-4 flex items-start gap-4">
            <Avatar className="size-16">
              {dev.avatar_url && <AvatarImage src={dev.avatar_url} alt={dev.github_username} />}
              <AvatarFallback className="text-lg">
                {dev.github_username[0]?.toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div className="min-w-0 flex-1">
              <h3 className="truncate text-lg font-bold">{dev.display_name ?? dev.github_username}</h3>
              <p className="font-mono text-sm text-muted-foreground">@{dev.github_username}</p>
              {dev.bio && <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">{dev.bio}</p>}
            </div>
          </div>
          <div className="mt-4 flex flex-wrap gap-2">
            <Badge variant="secondary">★ {dev.total_stars}</Badge>
            <Badge variant="secondary">{dev.public_repos} repos</Badge>
            <Badge variant="secondary">Score {Math.round(dev.activity_score)}</Badge>
          </div>
          <Link
            href={`/developers/${dev.github_username}`}
            className={cn(buttonVariants({ variant: "outline" }), "mt-4 w-full gap-1.5")}
          >
            View Dev Card <ArrowRight className="size-4" />
          </Link>
        </>
      )}
    </div>
  );
}
