"use client";

import { cn } from "@/lib/utils";

export function Sparkline({
  data,
  width = 72,
  height = 24,
  className,
  positive,
}: {
  data: { value: number }[];
  width?: number;
  height?: number;
  className?: string;
  positive?: boolean;
}) {
  if (!data.length) {
    return <div className={cn("text-muted-foreground", className)}>—</div>;
  }

  const values = data.map((d) => d.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;
  const pad = 2;
  const innerW = width - pad * 2;
  const innerH = height - pad * 2;

  const points = values
    .map((v, i) => {
      const x = pad + (i / Math.max(values.length - 1, 1)) * innerW;
      const y = pad + innerH - ((v - min) / range) * innerH;
      return `${x},${y}`;
    })
    .join(" ");

  const trendUp = values[values.length - 1] >= values[0];
  const stroke =
    positive === undefined
      ? "var(--primary)"
      : positive
        ? "hsl(142 76% 36%)"
        : "hsl(0 72% 51%)";

  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      className={cn("shrink-0", className)}
      aria-hidden
    >
      <polyline
        fill="none"
        stroke={stroke}
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        points={points}
        opacity={trendUp ? 1 : 0.85}
      />
    </svg>
  );
}

export function RankDelta({
  delta,
  period = 7,
}: {
  delta?: number | null;
  period?: 7 | 30;
}) {
  if (delta == null || delta === 0) {
    return (
      <span className="text-xs text-muted-foreground" title={`No rank change in ${period}d`}>
        —
      </span>
    );
  }

  const up = delta > 0;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-0.5 text-xs font-medium tabular-nums",
        up ? "text-emerald-600 dark:text-emerald-400" : "text-rose-600 dark:text-rose-400"
      )}
      title={`${up ? "Up" : "Down"} ${Math.abs(delta)} places in ${period}d`}
    >
      {up ? "↑" : "↓"}
      {Math.abs(delta)}
    </span>
  );
}
