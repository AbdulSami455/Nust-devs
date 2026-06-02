"use client";

import { useMemo } from "react";
import type { ContributionDay } from "@/lib/api";

const LEVELS = [
  "bg-muted",
  "bg-emerald-200 dark:bg-emerald-900/60",
  "bg-emerald-400 dark:bg-emerald-700",
  "bg-emerald-600 dark:bg-emerald-500",
  "bg-emerald-800 dark:bg-emerald-400",
];

function parseDate(s: string) {
  const [y, m, d] = s.split("-").map(Number);
  return new Date(y, m - 1, d);
}

function formatDate(d: Date) {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

export function ContributionHeatmap({ days }: { days: ContributionDay[] }) {
  const { weeks, maxCount, total } = useMemo(() => {
    const map = new Map(days.map((d) => [d.date, d.count]));
    const totalCount = days.reduce((s, d) => s + d.count, 0);
    const max = Math.max(...days.map((d) => d.count), 1);

    if (days.length === 0) return { weeks: [] as { date: string; count: number }[][], maxCount: 1, total: 0 };

    const sorted = [...days].sort((a, b) => a.date.localeCompare(b.date));
    const end = parseDate(sorted[sorted.length - 1].date);
    const start = new Date(end);
    start.setDate(start.getDate() - 364);
    start.setDate(start.getDate() - start.getDay());

    const grid: { date: string; count: number }[][] = [];
    let week: { date: string; count: number }[] = [];
    const cursor = new Date(start);

    while (cursor <= end) {
      const key = formatDate(cursor);
      week.push({ date: key, count: map.get(key) ?? 0 });
      if (cursor.getDay() === 6) {
        grid.push(week);
        week = [];
      }
      cursor.setDate(cursor.getDate() + 1);
    }
    if (week.length > 0) grid.push(week);

    return { weeks: grid, maxCount: max, total: totalCount };
  }, [days]);

  if (weeks.length === 0) return null;

  return (
    <section className="space-y-3">
      <div className="flex items-end justify-between gap-4">
        <h3 className="font-semibold">Contributions</h3>
        <p className="text-xs text-muted-foreground">{total.toLocaleString()} in the last year</p>
      </div>
      <div className="overflow-x-auto pb-1">
        <div className="inline-flex gap-[3px]">
          {weeks.map((week, wi) => (
            <div key={wi} className="flex flex-col gap-[3px]">
              {week.map((d) => {
                const level =
                  d.count === 0 ? 0 : Math.min(4, Math.ceil((d.count / maxCount) * 4));
                return (
                  <div
                    key={d.date}
                    title={`${d.date}: ${d.count} contributions`}
                    className={`size-[11px] rounded-[2px] ${LEVELS[level]}`}
                  />
                );
              })}
            </div>
          ))}
        </div>
      </div>
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <span>Less</span>
        {LEVELS.map((c, i) => (
          <div key={i} className={`size-[11px] rounded-[2px] ${c}`} />
        ))}
        <span>More</span>
      </div>
    </section>
  );
}
