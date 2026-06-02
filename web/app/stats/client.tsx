"use client";

import { useEffect, useState } from "react";
import {
  BarChart, Bar, XAxis, YAxis, Tooltip,
  PieChart, Pie, Cell, Legend,
} from "recharts";
import { api, type Overview, type LanguageStat, type Developer } from "@/lib/api";
import { ChartContainer } from "@/components/charts/chart-container";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const COLORS = ["#3b82f6","#10b981","#f59e0b","#ef4444","#8b5cf6","#06b6d4","#f97316","#ec4899","#14b8a6","#84cc16"];

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
      <h2 className="text-2xl font-bold">Platform Stats</h2>

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
              <BarChart data={devsChart} layout="vertical" margin={{ left: 16, right: 16 }}>
                <XAxis type="number" tick={{ fontSize: 11 }} />
                <YAxis type="category" dataKey="name" tick={{ fontSize: 11 }} width={90} />
                <Tooltip />
                <Bar dataKey="score" fill="#3b82f6" radius={[0, 4, 4, 0]} />
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
                <Pie data={langPie} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={100} label={({ name }) => name}>
                  {langPie.map((entry, i) => (
                    <Cell key={i} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip formatter={(v) => [`${v} repos`, "Repos"]} />
                <Legend />
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
              <BarChart data={langBar} margin={{ left: 8, right: 8 }}>
                <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                <YAxis tick={{ fontSize: 11 }} />
                <Tooltip formatter={(v) => [`${v} MB`, "Code"]} />
                <Bar dataKey="mb" radius={[4, 4, 0, 0]}>
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
