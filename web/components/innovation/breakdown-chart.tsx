"use client";

import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { NameCount, ContributorStat } from "@/lib/api";
import { ChartContainer } from "@/components/charts/chart-container";
import { Skeleton } from "@/components/ui/skeleton";

const COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
];

const CHART_HEIGHT = 208;
const CONTRIBUTORS_HEIGHT = 224;

export function BreakdownChart({
  title,
  subtitle,
  data,
  loading,
}: {
  title: string;
  subtitle?: string;
  data: NameCount[];
  loading: boolean;
}) {
  const chartData = data.map((d) => ({ name: d.name, count: d.count }));

  return (
    <div className="bento-card min-w-0">
      <div className="mb-4">
        <h3 className="font-semibold">{title}</h3>
        {subtitle && <p className="text-xs text-muted-foreground">{subtitle}</p>}
      </div>
      {loading ? (
        <Skeleton className="w-full" style={{ height: CHART_HEIGHT }} />
      ) : chartData.length === 0 ? (
        <div
          className="flex items-center justify-center text-sm text-muted-foreground"
          style={{ height: CHART_HEIGHT }}
        >
          No data yet.
        </div>
      ) : (
        <ChartContainer height={CHART_HEIGHT}>
          <BarChart data={chartData} layout="vertical" margin={{ left: 8, right: 8, top: 0, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" horizontal={false} />
            <XAxis type="number" tick={{ fontSize: 10 }} tickLine={false} axisLine={false} />
            <YAxis
              type="category"
              dataKey="name"
              tick={{ fontSize: 10 }}
              width={100}
              tickLine={false}
              axisLine={false}
            />
            <Tooltip
              contentStyle={{
                background: "var(--popover)",
                border: "1px solid var(--border)",
                borderRadius: "8px",
                fontSize: "12px",
              }}
            />
            <Bar dataKey="count" radius={[0, 4, 4, 0]}>
              {chartData.map((_, i) => (
                <Cell key={i} fill={COLORS[i % COLORS.length]} />
              ))}
            </Bar>
          </BarChart>
        </ChartContainer>
      )}
    </div>
  );
}

export function ContributorsChart({
  title,
  data,
  loading,
}: {
  title: string;
  data: ContributorStat[];
  loading: boolean;
}) {
  const chartData = data.map((d) => ({
    name: d.name || d.username,
    score: Math.round(d.score),
  }));

  return (
    <div className="bento-card min-w-0 lg:col-span-2">
      <div className="mb-4">
        <h3 className="font-semibold">{title}</h3>
        <p className="text-xs text-muted-foreground">Top tracked NUST developers by activity score</p>
      </div>
      {loading ? (
        <Skeleton className="w-full" style={{ height: CONTRIBUTORS_HEIGHT }} />
      ) : (
        <ChartContainer height={CONTRIBUTORS_HEIGHT}>
          <BarChart data={chartData} margin={{ left: 8, right: 8, top: 0, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" vertical={false} />
            <XAxis dataKey="name" tick={{ fontSize: 10 }} tickLine={false} axisLine={false} />
            <YAxis tick={{ fontSize: 10 }} tickLine={false} axisLine={false} width={36} />
            <Tooltip
              contentStyle={{
                background: "var(--popover)",
                border: "1px solid var(--border)",
                borderRadius: "8px",
                fontSize: "12px",
              }}
            />
            <Bar dataKey="score" fill="var(--chart-2)" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ChartContainer>
      )}
    </div>
  );
}
