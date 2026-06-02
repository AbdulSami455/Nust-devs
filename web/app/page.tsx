"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowRight, Code2 } from "lucide-react";
import { api, type Developer, type Overview, type PublicRepo, type CommunityActivityDay } from "@/lib/api";
import { BentoStats } from "@/components/dashboard/bento-stats";
import { ActivityChart } from "@/components/dashboard/activity-chart";
import { DeveloperSpotlight } from "@/components/dashboard/developer-spotlight";
import { DevCardMini } from "@/components/dashboard/dev-card-mini";
import { buttonVariants } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

export default function HomePage() {
  const [overview, setOverview] = useState<Overview | null>(null);
  const [activity, setActivity] = useState<CommunityActivityDay[]>([]);
  const [spotlight, setSpotlight] = useState<Developer | null>(null);
  const [topDevs, setTopDevs] = useState<Developer[]>([]);
  const [topProjects, setTopProjects] = useState<PublicRepo[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      api.public.overview().then(setOverview).catch(() => {}),
      api.public.communityActivity(30).then(setActivity).catch(() => {}),
      api.public.spotlight().then(setSpotlight).catch(() => {}),
      api.public.leaderboard("activity_score", 1, 6).then(setTopDevs).catch(() => {}),
      api.public.topProjects().then((p) => setTopProjects(p.slice(0, 5))).catch(() => {}),
    ]).finally(() => setLoading(false));
  }, []);

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 sm:py-12">
      {/* Hero */}
      <section className="mb-10 flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
        <div className="space-y-3">
          <Badge variant="secondary" className="gap-1">
            <Code2 className="size-3" /> NUST × GitHub
          </Badge>
          <h1 className="text-4xl font-bold tracking-tight sm:text-5xl">
            NUST Devs on <span className="text-gradient-gold">GitHub</span>
          </h1>
          <p className="max-w-xl text-muted-foreground">
            Track contributions, repositories, and top projects from NUST developers — community stats, leaderboards, and dev profiles in one place.
          </p>
        </div>
        <div className="flex gap-2">
          <Link href="/developers" className={cn(buttonVariants(), "gap-1.5")}>
            Explore Developers <ArrowRight className="size-4" />
          </Link>
          <Link href="/leaderboard" className={buttonVariants({ variant: "outline" })}>
            Leaderboard
          </Link>
        </div>
      </section>

      {/* Bento grid */}
      <section className="mb-10 space-y-4">
        <BentoStats overview={overview} loading={loading} />
      </section>

      <section className="mb-10 grid gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <ActivityChart data={activity} loading={loading} />
        </div>
        <DeveloperSpotlight dev={spotlight} loading={loading} />
      </section>

      {/* Top developers */}
      <section className="mb-10 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold">Top Developers</h2>
          <Link href="/developers" className="text-sm text-primary hover:underline">
            View all →
          </Link>
        </div>
        {loading ? (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-28 rounded-2xl" />
            ))}
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {topDevs.map((dev) => (
              <DevCardMini key={dev.id} dev={dev} />
            ))}
          </div>
        )}
      </section>

      {/* Top projects */}
      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold">Top Projects</h2>
          <Link href="/projects" className="text-sm text-primary hover:underline">
            View all →
          </Link>
        </div>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {loading
            ? Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-24 rounded-2xl" />)
            : topProjects.map((repo) => (
                <a
                  key={repo.id}
                  href={repo.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="bento-card block transition-colors hover:border-primary/40"
                >
                  <p className="truncate font-mono text-sm font-medium">{repo.full_name}</p>
                  <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">
                    {repo.description || "No description"}
                  </p>
                  <div className="mt-3 flex gap-2 text-xs text-muted-foreground">
                    <span>★ {repo.stars}</span>
                    {repo.language && <Badge variant="secondary" className="text-[10px]">{repo.language}</Badge>}
                  </div>
                </a>
              ))}
        </div>
      </section>
    </div>
  );
}
