"use client";

import { useEffect, useState } from "react";
import { api, type InnovationGraph, type ActivityEvent } from "@/lib/api";
import { TrendChart } from "@/components/innovation/trend-chart";
import { BreakdownChart, ContributorsChart } from "@/components/innovation/breakdown-chart";
import { LiveActivityFeed } from "@/components/dashboard/live-activity-feed";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

type Tab = "graph" | "activity";

export function InnovationClient() {
  const [tab, setTab] = useState<Tab>("graph");
  const [granularity, setGranularity] = useState<"quarterly" | "monthly">("quarterly");
  const [graph, setGraph] = useState<InnovationGraph | null>(null);
  const [activity, setActivity] = useState<ActivityEvent[]>([]);
  const [loadingGraph, setLoadingGraph] = useState(true);
  const [loadingActivity, setLoadingActivity] = useState(true);

  useEffect(() => {
    api.public
      .innovationGraph(granularity, granularity === "quarterly" ? 8 : 12)
      .then(setGraph)
      .catch(() => setGraph(null))
      .finally(() => setLoadingGraph(false));
  }, [granularity]);

  useEffect(() => {
    api.public
      .recentActivity(20)
      .then(setActivity)
      .catch(() => setActivity([]))
      .finally(() => setLoadingActivity(false));
  }, []);

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 sm:py-12">
      <section className="mb-8 space-y-4">
        <Badge variant="secondary">Ecosystem analytics</Badge>
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">NUST Innovation Graph</h1>
            <p className="mt-2 max-w-2xl text-muted-foreground">
              Quarterly trends and live activity from NUST developers on GitHub.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => setTab("graph")}
              className={cn(
                "rounded-lg px-3 py-1.5 text-sm font-medium transition-colors",
                tab === "graph" ? "bg-primary text-primary-foreground" : "bg-muted hover:bg-muted/80"
              )}
            >
              Innovation Graph
            </button>
            <button
              type="button"
              onClick={() => setTab("activity")}
              className={cn(
                "rounded-lg px-3 py-1.5 text-sm font-medium transition-colors",
                tab === "activity" ? "bg-primary text-primary-foreground" : "bg-muted hover:bg-muted/80"
              )}
            >
              Live Activity
            </button>
          </div>
        </div>
      </section>

      {tab === "activity" ? (
        <LiveActivityFeed events={activity} loading={loadingActivity} />
      ) : (
        <>
          <div className="mb-6 flex flex-wrap gap-2">
            <span className="self-center text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Period
            </span>
            {(["quarterly", "monthly"] as const).map((g) => (
              <button
                key={g}
                type="button"
                onClick={() => {
                  setLoadingGraph(true);
                  setGranularity(g);
                }}
                className={cn(
                  "rounded-full px-3 py-1 text-xs font-medium transition-colors",
                  granularity === g ? "bg-primary text-primary-foreground" : "bg-muted hover:bg-muted/80"
                )}
              >
                {g === "quarterly" ? "Quarterly" : "Monthly"}
              </button>
            ))}
          </div>

          {loadingGraph && !graph ? (
            <div className="grid min-w-0 gap-4 md:grid-cols-2">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-72 rounded-2xl" />
              ))}
            </div>
          ) : (
            <div className="grid min-w-0 gap-4 md:grid-cols-2">
              <TrendChart
                title="Git Pushes"
                subtitle="Contribution events across tracked developers"
                data={graph?.pushes ?? []}
                loading={loadingGraph}
                color="var(--chart-1)"
              />
              <TrendChart
                title="Repositories"
                subtitle="Distinct repos with push activity per period"
                data={graph?.repositories ?? []}
                loading={loadingGraph}
                color="var(--chart-3)"
              />
              <TrendChart
                title="Developers"
                subtitle="New developers added to the platform"
                data={graph?.developers ?? []}
                loading={loadingGraph}
                color="var(--chart-2)"
              />
              <TrendChart
                title="Organizations"
                subtitle="Developers with company/org affiliation added"
                data={graph?.organizations ?? []}
                loading={loadingGraph}
                color="var(--chart-4)"
              />
              <TrendChart
                title="Net New Stars"
                subtitle="Quarter-over-quarter star growth across tracked repos"
                data={graph?.net_new_stars ?? []}
                loading={loadingGraph}
                color="var(--chart-5)"
              />
              <BreakdownChart
                title="Programming Languages"
                subtitle="By code volume across tracked repos"
                data={graph?.languages ?? []}
                loading={loadingGraph}
              />
              <BreakdownChart
                title="Licenses"
                subtitle="SPDX licenses on tracked repositories"
                data={graph?.licenses ?? []}
                loading={loadingGraph}
              />
              <BreakdownChart
                title="Top Organizations"
                subtitle="Companies and affiliations among NUST devs"
                data={graph?.top_organizations ?? []}
                loading={loadingGraph}
              />
              <ContributorsChart
                title="Top Contributors"
                data={graph?.top_contributors ?? []}
                loading={loadingGraph}
              />
            </div>
          )}

          <p className="mt-8 text-center text-sm text-muted-foreground">
            License data populates after the next developer sync.{" "}
            <a href="/stats" className="text-primary hover:underline">
              View platform stats
            </a>
          </p>
        </>
      )}
    </div>
  );
}
