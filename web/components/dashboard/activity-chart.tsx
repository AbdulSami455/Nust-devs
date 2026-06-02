"use client";

import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { CommunityActivityDay } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function ActivityChart({
  data,
  loading,
}: {
  data: CommunityActivityDay[];
  loading: boolean;
}) {
  const chartData = data.map((d) => ({
    date: d.date.slice(5),
    commits: d.count,
  }));

  return (
    <div className="bento-card col-span-full lg:col-span-2 lg:row-span-2">
      <div className="mb-4">
        <h3 className="font-semibold">Community Pulse</h3>
        <p className="text-sm text-muted-foreground">Contributions across all tracked developers (30 days)</p>
      </div>
      {loading ? (
        <Skeleton className="h-64 w-full" />
      ) : chartData.length === 0 ? (
        <div className="flex h-64 items-center justify-center text-sm text-muted-foreground">
          No activity data yet — sync developers to populate the chart.
        </div>
      ) : (
        <div className="h-64 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData} margin={{ top: 4, right: 4, left: -20, bottom: 0 }}>
              <defs>
                <linearGradient id="pulse" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="var(--chart-1)" stopOpacity={0.4} />
                  <stop offset="100%" stopColor="var(--chart-1)" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" vertical={false} />
              <XAxis
                dataKey="date"
                tick={{ fontSize: 11 }}
                tickLine={false}
                axisLine={false}
                className="text-muted-foreground"
              />
              <YAxis tick={{ fontSize: 11 }} tickLine={false} axisLine={false} width={32} />
              <Tooltip
                contentStyle={{
                  background: "var(--popover)",
                  border: "1px solid var(--border)",
                  borderRadius: "8px",
                  fontSize: "12px",
                }}
              />
              <Area
                type="monotone"
                dataKey="commits"
                stroke="var(--chart-1)"
                strokeWidth={2}
                fill="url(#pulse)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}
