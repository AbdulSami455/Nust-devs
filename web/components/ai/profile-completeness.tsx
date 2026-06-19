"use client";

import type { Developer, PublicRepo } from "@/lib/api";
import { cn } from "@/lib/utils";
import { BadgeCheck, FileText, GitBranch, Globe, Mail, Sparkles } from "lucide-react";

type CompletenessItem = {
  label: string;
  score: number;
  met: boolean;
  note: string;
};

function normalizeEmail(email?: string) {
  return email?.trim().toLowerCase() ?? "";
}

function hasRecentActivity(repo: PublicRepo) {
  if (!repo.pushed_at) return false;
  const pushedAt = new Date(repo.pushed_at);
  if (Number.isNaN(pushedAt.getTime())) return false;
  const ageDays = (Date.now() - pushedAt.getTime()) / (1000 * 60 * 60 * 24);
  return ageDays <= 90;
}

function clampScore(value: number) {
  return Math.max(0, Math.min(100, Math.round(value)));
}

function buildItems(dev: Developer, repos: PublicRepo[]): CompletenessItem[] {
  const originalRepos = repos.filter((repo) => !repo.is_fork);
  const activeRepos = originalRepos.filter(hasRecentActivity);
  const readmeRepos = dev.readme_repos ?? 0;
  const email = normalizeEmail(dev.email);
  const hasNustEmail = email.endsWith("@nust.edu.pk");

  const readmeCoverage = originalRepos.length > 0 ? Math.min(readmeRepos / originalRepos.length, 1) : 0;
  const activityCoverage = originalRepos.length > 0 ? Math.min(activeRepos.length / Math.max(1, Math.min(originalRepos.length, 4)), 1) : 0;

  return [
    {
      label: "Bio",
      score: dev.bio?.trim() ? 20 : 0,
      met: Boolean(dev.bio?.trim()),
      note: dev.bio?.trim() ? "Bio is set." : "Missing profile bio.",
    },
    {
      label: "Website",
      score: dev.website?.trim() ? 15 : 0,
      met: Boolean(dev.website?.trim()),
      note: dev.website?.trim() ? "Website is set." : "Missing personal or project website.",
    },
    {
      label: "NUST email",
      score: hasNustEmail ? 15 : 0,
      met: hasNustEmail,
      note: hasNustEmail ? "NUST email is present." : "Missing verified NUST email.",
    },
    {
      label: "Active repos",
      score: activityCoverage * 25,
      met: activeRepos.length > 0,
      note:
        activeRepos.length > 0
          ? `${activeRepos.length} active repo${activeRepos.length === 1 ? "" : "s"} in the last 90 days.`
          : "No recently active repos found.",
    },
    {
      label: "README coverage",
      score: readmeCoverage * 25,
      met: readmeRepos > 0,
      note:
        originalRepos.length > 0
          ? `${readmeRepos}/${originalRepos.length} original repo${originalRepos.length === 1 ? "" : "s"} with README data.`
          : "No original repos tracked yet.",
    },
  ];
}

export function ProfileCompletenessCard({
  dev,
  repos,
}: {
  dev: Developer;
  repos: PublicRepo[];
}) {
  const items = buildItems(dev, repos);
  const score = clampScore(items.reduce((sum, item) => sum + item.score, 0));
  const missing = items.filter((item) => !item.met);

  const status =
    score >= 85 ? "Strong profile signals" :
    score >= 65 ? "Moderate profile signals" :
    "Incomplete profile signals";

  return (
    <section className="bento-card space-y-4">
      <div className="flex items-start justify-between gap-3">
        <div className="space-y-1">
          <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            <Sparkles className="size-3.5" />
            AI profile completeness
          </div>
          <h2 className="text-lg font-semibold">{status}</h2>
        </div>
        <div className="text-right">
          <p className="text-3xl font-semibold tabular-nums">{score}</p>
          <p className="text-xs text-muted-foreground">/ 100</p>
        </div>
      </div>

      <div className="h-2 overflow-hidden rounded-full bg-muted">
        <div
          className={cn(
            "h-full rounded-full transition-all",
            score >= 85 ? "bg-emerald-500" : score >= 65 ? "bg-amber-500" : "bg-primary",
          )}
          style={{ width: `${score}%` }}
        />
      </div>

      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
        {items.map((item) => (
          <div key={item.label} className="rounded-xl border border-border/70 bg-background/60 p-3">
            <div className="flex items-center gap-2">
              {item.label === "Bio" ? <FileText className="size-4 text-muted-foreground" /> :
               item.label === "Website" ? <Globe className="size-4 text-muted-foreground" /> :
               item.label === "NUST email" ? <Mail className="size-4 text-muted-foreground" /> :
               item.label === "Active repos" ? <GitBranch className="size-4 text-muted-foreground" /> :
               <BadgeCheck className="size-4 text-muted-foreground" />}
              <p className="text-sm font-medium">{item.label}</p>
            </div>
            <p className="mt-2 text-sm text-muted-foreground">{item.note}</p>
          </div>
        ))}
      </div>

      <div className="flex flex-wrap gap-2">
        {missing.length === 0 ? (
          <span className="inline-flex items-center rounded-full border border-emerald-500/30 bg-emerald-500/10 px-3 py-1 text-xs text-emerald-700 dark:text-emerald-300">
            All core profile signals present
          </span>
        ) : (
          missing.map((item) => (
            <span
              key={item.label}
              className="inline-flex items-center rounded-full border border-border bg-muted px-3 py-1 text-xs text-muted-foreground"
            >
              Missing {item.label.toLowerCase()}
            </span>
          ))
        )}
      </div>
    </section>
  );
}
