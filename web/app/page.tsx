"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { ArrowRight, GitPullRequest, ShieldCheck } from "lucide-react";
import {
  api,
  type Developer,
  type Overview,
  type PublicRepo,
  type CommunityActivityDay,
  type ActivityEvent,
  type OSSStats,
  type StreakSummary,
} from "@/lib/api";
import { StreakWidget } from "@/components/dashboard/streak-widget";
import { BentoStats } from "@/components/dashboard/bento-stats";
import { ActivityChart } from "@/components/dashboard/activity-chart";
import { DeveloperSpotlight } from "@/components/dashboard/developer-spotlight";
import { LiveActivityFeed } from "@/components/dashboard/live-activity-feed";
import { OSSStatsPanel } from "@/components/dashboard/oss-stats";
import { DevCardMini } from "@/components/dashboard/dev-card-mini";
import { ProjectImpactSummaryCard } from "@/components/ai/project-summary";
import { NormalizedTagsCard } from "@/components/ai/normalized-tags";
import { buttonVariants } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

function asArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

export default function HomePage() {
  const [overview, setOverview] = useState<Overview | null>(null);
  const [activity, setActivity] = useState<CommunityActivityDay[]>([]);
  const [spotlight, setSpotlight] = useState<Developer | null>(null);
  const [topDevs, setTopDevs] = useState<Developer[]>([]);
  const [topProjects, setTopProjects] = useState<PublicRepo[]>([]);
  const [recentEvents, setRecentEvents] = useState<ActivityEvent[]>([]);
  const [ossStats, setOssStats] = useState<OSSStats | null>(null);
  const [streakSummary, setStreakSummary] = useState<StreakSummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      api.public.overview().then(setOverview).catch(() => {}),
      api.public.communityActivity(30).then((data) => setActivity(asArray(data))).catch(() => {}),
      api.public.spotlight().then(setSpotlight).catch(() => {}),
      api.public.leaderboard("activity_score", 1, 6).then((data) => setTopDevs(asArray(data))).catch(() => {}),
      api.public.topProjects({ category: "original", limit: 5 }).then((data) => setTopProjects(asArray(data))).catch(() => {}),
      api.public.recentActivity(12).then((data) => setRecentEvents(asArray(data))).catch(() => {}),
      api.public.openSource().then(setOssStats).catch(() => {}),
      api.public.streakSummary().then(setStreakSummary).catch(() => {}),
    ]).finally(() => setLoading(false));
  }, []);

  const contributionCount = overview?.total_contributions ?? 0;

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 sm:py-12">
      <section className="mb-12 grid gap-8 lg:grid-cols-[minmax(0,1fr)_360px] lg:items-end">
        <div className="max-w-3xl">
          <Badge variant="secondary" className="mb-5 gap-2 rounded-full px-3 py-1">
            <ShieldCheck className="size-3.5" />
            NUST developer community
          </Badge>
          <h1 className="max-w-3xl text-4xl font-semibold leading-[1.04] tracking-tight text-foreground sm:text-6xl">
            A cleaner public record of NUST builders on GitHub.
          </h1>
          <p className="mt-5 max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg">
            Discover student developers, follow open-source momentum, and compare real activity
            across repositories, languages, streaks, and leaderboards.
          </p>
          <div className="mt-7 flex flex-wrap gap-3">
            <Link href="/developers" className={cn(buttonVariants(), "gap-2")}>
              Explore developers
              <ArrowRight className="size-4" />
            </Link>
            <Link href="/join" className={cn(buttonVariants({ variant: "outline" }), "gap-2")}>
              Join as developer
              <GitPullRequest className="size-4" />
            </Link>
          </div>
        </div>

        <div className="bento-card relative overflow-hidden p-0">
          <div className="border-b border-border/80 bg-secondary/40 px-5 py-4">
            <div className="flex items-center gap-3">
              <Image
                src="/nust-logo.svg"
                alt="NUST"
                width={48}
                height={48}
                className="size-12 rounded-full ring-1 ring-border"
              />
              <div>
                <p className="text-sm font-semibold">NUST Devs Index</p>
                <p className="text-xs text-muted-foreground">GitHub activity snapshot</p>
              </div>
            </div>
          </div>
          <div className="grid grid-cols-2 divide-x divide-y divide-border/70">
            <div className="p-5">
              <p className="text-xs font-medium uppercase text-muted-foreground">Developers</p>
              <p className="mt-2 text-3xl font-semibold tabular-nums">
                {(overview?.total_developers ?? 0).toLocaleString()}
              </p>
            </div>
            <div className="p-5">
              <p className="text-xs font-medium uppercase text-muted-foreground">Repositories</p>
              <p className="mt-2 text-3xl font-semibold tabular-nums">
                {(overview?.total_repos ?? 0).toLocaleString()}
              </p>
            </div>
            <div className="p-5">
              <p className="text-xs font-medium uppercase text-muted-foreground">Stars</p>
              <p className="mt-2 text-3xl font-semibold tabular-nums">
                {(overview?.total_stars ?? 0).toLocaleString()}
              </p>
            </div>
            <div className="p-5">
              <p className="text-xs font-medium uppercase text-muted-foreground">Contributions</p>
              <p className="mt-2 text-3xl font-semibold tabular-nums">
                {contributionCount.toLocaleString()}
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="mb-10">
        <BentoStats overview={overview} loading={loading} />
      </section>

      <section className="mb-10">
        <StreakWidget summary={streakSummary} loading={loading} />
      </section>

      <section className="mb-10">
        <OSSStatsPanel stats={ossStats} loading={loading} />
      </section>

      <section className="mb-10 grid gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <ActivityChart data={activity} loading={loading} />
        </div>
        <DeveloperSpotlight dev={spotlight} loading={loading} />
      </section>

      <section className="mb-10">
        <LiveActivityFeed events={recentEvents} loading={loading} />
      </section>

      <section className="mb-10 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold tracking-tight">Top Developers</h2>
          <Link href="/developers" className="text-sm text-primary hover:underline">
            View all
          </Link>
        </div>
        {loading ? (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-28 rounded-lg" />
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

      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold tracking-tight">Top Open Source Projects</h2>
          <Link href="/projects" className="text-sm text-primary hover:underline">
            View all
          </Link>
        </div>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {loading
            ? Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-24 rounded-lg" />)
            : topProjects.map((repo) => (
                <div key={repo.id}>
                  <a
                    href={repo.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="bento-card block transition-colors hover:border-primary/40"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <p className="truncate font-mono text-sm font-medium">{repo.full_name}</p>
                      {!repo.is_fork && (
                        <Badge variant="outline" className="shrink-0 text-[10px]">
                          OSS
                        </Badge>
                      )}
                    </div>
                    <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">
                      {repo.description || "No description"}
                    </p>
                    <div className="mt-3 flex gap-2 text-xs text-muted-foreground">
                      <span>{repo.stars} stars</span>
                      {repo.language && (
                        <Badge variant="secondary" className="text-[10px]">
                          {repo.language}
                        </Badge>
                      )}
                    </div>
                  </a>
                  <ProjectImpactSummaryCard repo={repo} />
                  <NormalizedTagsCard kind="project" repoId={repo.id} />
                </div>
              ))}
        </div>
      </section>
    </div>
  );
}
