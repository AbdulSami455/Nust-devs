"use client";

import type { ContributionStats } from "@/lib/api";
import { Badge } from "@/components/ui/badge";

export function ContributionStatsPanel({ stats }: { stats: ContributionStats }) {
  const hasTotals =
    stats.pull_requests > 0 || stats.issues > 0 || stats.reviews > 0;
  const hasRepos = stats.by_repository.length > 0;

  if (!hasTotals && !hasRepos) {
    return null;
  }

  const periodLabel =
    stats.period_start && stats.period_end
      ? `${stats.period_start} → ${stats.period_end}`
      : "Last 12 months";

  return (
    <section className="bento-card space-y-5">
      <div>
        <h2 className="text-lg font-semibold">Open source contributions</h2>
        <p className="text-sm text-muted-foreground">
          PRs, issues, and reviews from GitHub GraphQL · {periodLabel}
        </p>
      </div>

      {hasTotals && (
        <div className="grid grid-cols-3 gap-3">
          {[
            { label: "Pull requests", value: stats.pull_requests },
            { label: "Issues", value: stats.issues },
            { label: "Reviews", value: stats.reviews },
          ].map(({ label, value }) => (
            <div key={label} className="rounded-xl border bg-background/60 px-4 py-3">
              <p className="text-xs text-muted-foreground">{label}</p>
              <p className="text-2xl font-bold tabular-nums">{value.toLocaleString()}</p>
            </div>
          ))}
        </div>
      )}

      {hasRepos && (
        <div className="space-y-3">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Top repositories contributed to
          </p>
          <div className="overflow-hidden rounded-xl border">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/40 text-left text-xs text-muted-foreground">
                  <th className="px-4 py-2 font-medium">Repository</th>
                  <th className="hidden px-4 py-2 font-medium sm:table-cell">PRs</th>
                  <th className="hidden px-4 py-2 font-medium md:table-cell">Issues</th>
                  <th className="hidden px-4 py-2 font-medium lg:table-cell">Reviews</th>
                  <th className="px-4 py-2 font-medium">Total</th>
                </tr>
              </thead>
              <tbody>
                {stats.by_repository.slice(0, 15).map((repo) => (
                  <tr key={repo.repo_full_name} className="border-b last:border-0">
                    <td className="px-4 py-2.5">
                      <a
                        href={repo.repo_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="font-mono text-xs hover:text-primary hover:underline"
                      >
                        {repo.repo_full_name}
                      </a>
                    </td>
                    <td className="hidden px-4 py-2.5 tabular-nums sm:table-cell">
                      {repo.pull_requests || "—"}
                    </td>
                    <td className="hidden px-4 py-2.5 tabular-nums md:table-cell">
                      {repo.issues || "—"}
                    </td>
                    <td className="hidden px-4 py-2.5 tabular-nums lg:table-cell">
                      {repo.reviews || "—"}
                    </td>
                    <td className="px-4 py-2.5">
                      <Badge variant="secondary" className="tabular-nums">
                        {repo.total}
                      </Badge>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </section>
  );
}
