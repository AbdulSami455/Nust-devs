"use client";

import {
  Area,
  AreaChart,
  CartesianGrid,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { TrendPoint } from "@/lib/api";
import { ChartContainer } from "@/components/charts/chart-container";
import { Skeleton } from "@/components/ui/skeleton";

const CHART_HEIGHT = 208;

export function TrendChart({
  title,
  subtitle,
  data,
  loading,
  color = "var(--chart-1)",
}: {
  title: string;
  subtitle?: string;
  data: TrendPoint[];
  loading: boolean;
  color?: string;
}) {
  const chartData = data.map((d) => ({ label: d.label, value: d.value }));
  const gradientId = `trend-${title.replace(/[^a-zA-Z0-9]/g, "-").toLowerCase()}`;

  return (
    <div className="bento-card min-w-0">
      <div className="mb-4">
        <h3 className="font-semibold">{title}</h3>
        {subtitle && <p className="text-xs text-muted-foreground">{subtitle}</p>}
      </div>
      {loading ? (
        <Skeleton className="w-full" style={{ height: CHART_HEIGHT }} />
      ) : chartData.every((d) => d.value === 0) ? (
        <div
          className="flex items-center justify-center text-sm text-muted-foreground"
          style={{ height: CHART_HEIGHT }}
        >
          No data for this period yet.
        </div>
      ) : (
        <ChartContainer height={CHART_HEIGHT}>
          <AreaChart data={chartData} margin={{ top: 4, right: 4, left: -16, bottom: 0 }}>
            <defs>
              <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={color} stopOpacity={0.35} />
                <stop offset="100%" stopColor={color} stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" vertical={false} />
            <XAxis
              dataKey="label"
              tick={{ fontSize: 10 }}
              tickLine={false}
              axisLine={false}
              interval="preserveStartEnd"
            />
            <YAxis tick={{ fontSize: 10 }} tickLine={false} axisLine={false} width={36} />
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
              dataKey="value"
              stroke={color}
              strokeWidth={2}
              fill={`url(#${gradientId})`}
            />
          </AreaChart>
        </ChartContainer>
      )}
    </div>
  );
}
