"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, type DevOfMonthWinner } from "@/lib/api";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { powerTitle } from "@/components/profile/power-level";

const MONTHS = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
];

export default function HallOfFamePage() {
  const [winners, setWinners] = useState<DevOfMonthWinner[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.public
      .devOfMonth(24)
      .then(setWinners)
      .catch(() => setWinners([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="mx-auto max-w-4xl space-y-8 px-4 py-8 sm:px-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-bold tracking-tight">Hall of Fame</h1>
        <p className="text-sm text-muted-foreground">
          NUST Dev of the Month — scored by monthly activity, rank improvement, and stars gained.
        </p>
      </div>

      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-24 rounded-2xl" />
          ))}
        </div>
      ) : winners.length === 0 ? (
        <div className="rounded-2xl border border-dashed px-6 py-16 text-center text-muted-foreground">
          No winners yet. The worker awards Dev of the Month on the 1st of each month.
        </div>
      ) : (
        <div className="space-y-4">
          {winners.map((w) => {
            const d = w.developer;
            return (
              <div key={`${w.year}-${w.month}`} className="bento-card flex gap-4">
                <Avatar className="size-14 shrink-0">
                  {d.avatar_url && <AvatarImage src={d.avatar_url} alt={d.github_username} />}
                  <AvatarFallback>{d.github_username[0]?.toUpperCase()}</AvatarFallback>
                </Avatar>
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge className="bg-[var(--nust-gold)] text-black">
                      {MONTHS[w.month - 1]} {w.year}
                    </Badge>
                    <Link
                      href={`/developers/${d.github_username}`}
                      className="font-semibold hover:underline"
                    >
                      {d.display_name ?? d.github_username}
                    </Link>
                    <Badge variant="outline">
                      Lv.{d.power_level ?? 1} {powerTitle(d.power_level ?? 1)}
                    </Badge>
                  </div>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Score {Math.round(w.score)} · {w.activity_points} activity pts ·{" "}
                    {w.rank_gain >= 0 ? "+" : ""}
                    {w.rank_gain} rank · +{w.stars_gained} stars
                  </p>
                  <div className="mt-2 flex flex-wrap gap-2">
                    <Link
                      href={`/developers/${d.github_username}/wrapped?year=${w.year}`}
                      className="text-xs text-primary hover:underline"
                    >
                      View {w.year} wrapped
                    </Link>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
