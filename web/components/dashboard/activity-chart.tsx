"use client";

import {
  Area,
  AreaChart,
  CartesianGrid,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { CommunityActivityDay } from "@/lib/api";
import { ChartContainer } from "@/components/charts/chart-container";
import { Skeleton } from "@/components/ui/skeleton";

const CHART_HEIGHT = 256;

export function ActivityChart({
  data,
  loading,
}: {
  data: CommunityActivityDay[] | null | undefined;
  loading: boolean;
}) {
  const chartData = (Array.isArray(data) ? data : []).map((d) => ({
    date: d.date.slice(5),
    commits: d.count,
  }));

  return (
    <div className="bento-card col-span-full min-w-0 lg:col-span-2 lg:row-span-2">
      <div className="mb-4">
        <h3 className="text-base font-semibold tracking-tight">Community Pulse</h3>
        <p className="text-sm leading-6 text-muted-foreground">Contributions across all tracked developers over the last 30 days.</p>
      </div>
      {loading ? (
        <Skeleton className="w-full" style={{ height: CHART_HEIGHT }} />
      ) : chartData.length === 0 ? (
        <div
          className="flex items-center justify-center text-sm text-muted-foreground"
          style={{ height: CHART_HEIGHT }}
        >
          No activity data yet — sync developers to populate the chart.
        </div>
      ) : (
        <ChartContainer height={CHART_HEIGHT}>
          <AreaChart data={chartData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
            <defs>
              <linearGradient id="pulse" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="var(--chart-1)" stopOpacity={0.4} />
                <stop offset="100%" stopColor="var(--chart-1)" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="4 4" className="stroke-border/55" vertical={false} />
            <XAxis
              dataKey="date"
              tick={{ fontSize: 12, fontWeight: 500 }}
              tickLine={false}
              axisLine={false}
              className="text-muted-foreground"
            />
            <YAxis tick={{ fontSize: 12, fontWeight: 500 }} tickLine={false} axisLine={false} width={40} />
            <Tooltip
              cursor={{ stroke: "var(--border)", strokeWidth: 1 }}
              contentStyle={{
                background: "var(--popover)",
                border: "1px solid var(--border)",
                borderRadius: "8px",
                fontSize: "12px",
              }}
              wrapperClassName="chart-tooltip"
              formatter={(value) => [Number(value).toLocaleString(), "Contributions"]}
            />
            <Area
              type="monotone"
              dataKey="commits"
              stroke="var(--chart-1)"
              strokeWidth={2}
              fill="url(#pulse)"
            />
          </AreaChart>
        </ChartContainer>
      )}
    </div>
  );
}
