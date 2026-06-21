"use client";

import { useEffect, useState } from "react";
import {
  BarChart, Bar, XAxis, YAxis, Tooltip,
  PieChart, Pie, Cell, Legend, CartesianGrid,
} from "recharts";
import { api, type Overview, type LanguageStat, type Developer } from "@/lib/api";
import { ChartContainer } from "@/components/charts/chart-container";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PlatformInsightsCard } from "@/components/ai/platform-insights";
import { WeeklyCommunityReportCard } from "@/components/ai/weekly-community-report";

const COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
  "oklch(0.6 0.08 248)",
  "oklch(0.62 0.09 140)",
  "oklch(0.63 0.1 330)",
  "oklch(0.64 0.1 62)",
  "oklch(0.58 0.08 205)",
];

export function StatsClient() {
  const [overview, setOverview]   = useState<Overview | null>(null);
  const [languages, setLanguages] = useState<LanguageStat[]>([]);
  const [topDevs, setTopDevs]     = useState<Developer[]>([]);
  const [loading, setLoading]     = useState(true);

  useEffect(() => {
    Promise.all([
      api.public.overview(),
      api.public.languages(),
      api.public.leaderboard("activity_score", 1, 10),
    ]).then(([ov, langs, devs]) => {
      setOverview(ov);
      setLanguages(langs ?? []);
      setTopDevs(devs ?? []);
    }).catch(() => {}).finally(() => setLoading(false));
  }, []);

  if (loading) return <main className="p-8 text-muted-foreground">Loading…</main>;

  const langBar = languages.slice(0, 10).map((l) => ({
    name: l.language,
    repos: l.repo_count,
    mb: Math.round(l.bytes / 1024 / 1024),
  }));

  const langPie = languages.slice(0, 8).map((l, i) => ({
    name: l.language,
    value: l.repo_count,
    color: COLORS[i % COLORS.length],
  }));

  const devsChart = topDevs.map((d) => ({
    name: d.github_username,
    score: Math.round(d.activity_score),
    stars: d.total_stars,
  }));

  return (
    <main className="mx-auto max-w-6xl px-6 py-8 space-y-8">
      <div>
        <h2 className="text-2xl font-semibold tracking-tight">Platform Stats</h2>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          Aggregated GitHub activity, language usage, and developer rankings across tracked profiles.
        </p>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <PlatformInsightsCard />
        <WeeklyCommunityReportCard />
      </div>

      {/* Overview cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        {[
          { label: "Developers",    value: overview?.total_developers ?? 0 },
          { label: "Repositories",  value: overview?.total_repos ?? 0 },
          { label: "Total Stars",   value: overview?.total_stars ?? 0 },
          { label: "Contributions", value: (overview?.total_contributions ?? 0).toLocaleString() },
        ].map(({ label, value }) => (
          <Card key={label}>
            <CardHeader className="pb-1">
              <CardTitle className="text-sm font-medium text-muted-foreground">{label}</CardTitle>
            </CardHeader>
            <CardContent><p className="text-2xl font-bold">{value}</p></CardContent>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top developers by activity */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Top Developers — Activity Score</CardTitle>
          </CardHeader>
          <CardContent>
            <ChartContainer height={280}>
              <BarChart data={devsChart} layout="vertical" margin={{ left: 8, right: 16 }}>
                <CartesianGrid strokeDasharray="4 4" className="stroke-border/55" horizontal={false} />
                <XAxis type="number" tick={{ fontSize: 12, fontWeight: 500 }} tickLine={false} axisLine={false} />
                <YAxis type="category" dataKey="name" tick={{ fontSize: 12, fontWeight: 500 }} tickLine={false} axisLine={false} width={108} />
                <Tooltip wrapperClassName="chart-tooltip" formatter={(value) => [Number(value).toLocaleString(), "Score"]} />
                <Bar dataKey="score" fill="var(--chart-1)" radius={[0, 5, 5, 0]} />
              </BarChart>
            </ChartContainer>
          </CardContent>
        </Card>

        {/* Language distribution pie */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Language Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            <ChartContainer height={280}>
              <PieChart>
                <Pie data={langPie} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={94} labelLine={false} label={({ name }) => name}>
                  {langPie.map((entry, i) => (
                    <Cell key={i} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip wrapperClassName="chart-tooltip" formatter={(v) => [`${v} repos`, "Repos"]} />
                <Legend iconType="circle" />
              </PieChart>
            </ChartContainer>
          </CardContent>
        </Card>

        {/* Language bytes bar */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">Languages by Code Volume (MB)</CardTitle>
          </CardHeader>
          <CardContent>
            <ChartContainer height={240}>
              <BarChart data={langBar} margin={{ left: 8, right: 12 }}>
                <CartesianGrid strokeDasharray="4 4" className="stroke-border/55" vertical={false} />
                <XAxis dataKey="name" tick={{ fontSize: 12, fontWeight: 500 }} tickLine={false} axisLine={false} />
                <YAxis tick={{ fontSize: 12, fontWeight: 500 }} tickLine={false} axisLine={false} width={42} />
                <Tooltip wrapperClassName="chart-tooltip" formatter={(v) => [`${v} MB`, "Code"]} />
                <Bar dataKey="mb" radius={[5, 5, 0, 0]}>
                  {langBar.map((_, i) => (
                    <Cell key={i} fill={COLORS[i % COLORS.length]} />
                  ))}
                </Bar>
              </BarChart>
            </ChartContainer>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}
