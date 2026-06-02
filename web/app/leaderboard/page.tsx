"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, type Developer } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const SORT_OPTIONS = [
  { value: "activity_score", label: "Activity Score" },
  { value: "total_stars",    label: "Stars" },
  { value: "public_repos",   label: "Repos" },
  { value: "followers",      label: "Followers" },
];

export default function LeaderboardPage() {
  const [developers, setDevelopers] = useState<Developer[]>([]);
  const [sortBy, setSortBy] = useState("activity_score");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.public.leaderboard(sortBy, 1, 50).then(setDevelopers).catch(() => {}).finally(() => setLoading(false));
  }, [sortBy]);

  return (
    <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6 space-y-6">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold">Leaderboard</h2>
          <div className="flex gap-2">
            {SORT_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                onClick={() => setSortBy(opt.value)}
                className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                  sortBy === opt.value
                    ? "bg-primary text-primary-foreground"
                    : "bg-background border hover:bg-muted"
                }`}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>

        <div className="rounded-lg border bg-background overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-12">Rank</TableHead>
                <TableHead>Developer</TableHead>
                <TableHead>Activity Score</TableHead>
                <TableHead>Stars</TableHead>
                <TableHead>Repos</TableHead>
                <TableHead>Followers</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center text-muted-foreground py-8">Loading…</TableCell>
                </TableRow>
              ) : developers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center text-muted-foreground py-8">No developers synced yet.</TableCell>
                </TableRow>
              ) : (
                developers.map((dev, i) => (
                  <TableRow key={dev.id}>
                    <TableCell className="font-bold text-muted-foreground">
                      {i === 0 ? "🥇" : i === 1 ? "🥈" : i === 2 ? "🥉" : `#${i + 1}`}
                    </TableCell>
                    <TableCell>
                      <Link href={`/developers/${dev.github_username}`} className="flex items-center gap-2 hover:underline">
                        {dev.avatar_url && (
                          <img src={dev.avatar_url} alt={dev.github_username} className="w-7 h-7 rounded-full" />
                        )}
                        <div>
                          <p className="font-medium">{dev.display_name ?? dev.github_username}</p>
                          <p className="text-xs text-muted-foreground font-mono">@{dev.github_username}</p>
                        </div>
                      </Link>
                    </TableCell>
                    <TableCell className="font-mono font-medium">{Math.round(dev.activity_score)}</TableCell>
                    <TableCell>{dev.total_stars}</TableCell>
                    <TableCell>{dev.public_repos}</TableCell>
                    <TableCell>{dev.followers}</TableCell>
                    <TableCell>
                      <Badge variant={dev.verification_status === "registered" ? "secondary" : "default"} className="text-xs">
                        {dev.verification_status}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
    </div>
  );
}
