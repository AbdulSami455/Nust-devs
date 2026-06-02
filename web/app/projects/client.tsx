"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { formatDistanceToNow } from "date-fns";
import { api, type PublicRepo, type ProjectCategory, type ProjectSort } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

const LANG_COLORS: Record<string, string> = {
  Go: "bg-sky-400",
  TypeScript: "bg-blue-500",
  JavaScript: "bg-yellow-400",
  Python: "bg-green-500",
  Rust: "bg-orange-500",
  Java: "bg-red-500",
  "C++": "bg-pink-500",
  C: "bg-gray-500",
};

const CATEGORIES: { value: ProjectCategory; label: string }[] = [
  { value: "all", label: "All" },
  { value: "original", label: "Original OSS" },
  { value: "forks", label: "Forks" },
];

const SORTS: { value: ProjectSort; label: string }[] = [
  { value: "stars", label: "Most stars" },
  { value: "recent", label: "Recently updated" },
  { value: "forks", label: "Most forked" },
];

function FilterPill({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "rounded-full px-3 py-1 text-xs font-medium transition-colors",
        active ? "bg-primary text-primary-foreground" : "bg-muted hover:bg-muted/80"
      )}
    >
      {children}
    </button>
  );
}

export function ProjectsClient() {
  const [repos, setRepos] = useState<PublicRepo[]>([]);
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState<ProjectCategory>("original");
  const [sort, setSort] = useState<ProjectSort>("stars");
  const [lang, setLang] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.public
      .topProjects({ category, sort, language: lang || undefined, limit: 60 })
      .then(setRepos)
      .catch(() => setRepos([]))
      .finally(() => setLoading(false));
  }, [category, sort, lang]);

  const languages = Array.from(
    new Set(repos.map((r) => r.language).filter(Boolean) as string[])
  ).sort();

  const filtered = repos.filter((r) => {
    const q = query.toLowerCase();
    return (
      r.full_name.toLowerCase().includes(q) ||
      r.description?.toLowerCase().includes(q) ||
      r.language?.toLowerCase().includes(q)
    );
  });

  return (
    <div className="mx-auto max-w-6xl space-y-6 px-4 py-8 sm:px-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Open Source Projects</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Repositories from tracked NUST developers — filter by type, language, or activity.
          </p>
        </div>
        <Input
          placeholder="Search repos…"
          className="max-w-xs"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
      </div>

      <div className="space-y-3">
        <div>
          <p className="mb-2 text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Category
          </p>
          <div className="flex flex-wrap gap-2">
            {CATEGORIES.map(({ value, label }) => (
              <FilterPill
                key={value}
                active={category === value}
                onClick={() => setCategory(value)}
              >
                {label}
              </FilterPill>
            ))}
          </div>
        </div>

        <div>
          <p className="mb-2 text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Sort by
          </p>
          <div className="flex flex-wrap gap-2">
            {SORTS.map(({ value, label }) => (
              <FilterPill key={value} active={sort === value} onClick={() => setSort(value)}>
                {label}
              </FilterPill>
            ))}
          </div>
        </div>

        {languages.length > 0 && (
          <div>
            <p className="mb-2 text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Language
            </p>
            <div className="flex flex-wrap gap-2">
              <FilterPill active={lang === ""} onClick={() => setLang("")}>
                All languages
              </FilterPill>
              {languages.map((l) => (
                <FilterPill key={l} active={lang === l} onClick={() => setLang(l)}>
                  {l}
                </FilterPill>
              ))}
            </div>
          </div>
        )}
      </div>

      {loading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-36 rounded-2xl" />
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <div className="rounded-2xl border border-dashed px-6 py-16 text-center">
          <p className="font-medium">No repositories match your filters</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Try a different category or clear the search query.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((repo) => (
            <div
              key={repo.id}
              className="bento-card flex h-full flex-col transition-colors hover:border-primary/40"
            >
              <a
                href={repo.url}
                target="_blank"
                rel="noopener noreferrer"
                className="block flex-1"
              >
                <div className="flex items-start justify-between gap-2">
                  <p className="font-mono text-sm font-medium leading-snug">{repo.full_name}</p>
                  {repo.language && (
                    <span
                      className={`mt-1 size-2.5 shrink-0 rounded-full ${LANG_COLORS[repo.language] ?? "bg-muted-foreground"}`}
                    />
                  )}
                </div>
                <p className="mt-2 line-clamp-2 text-xs text-muted-foreground">
                  {repo.description || "No description"}
                </p>
                <div className="mt-4 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                  <span>{repo.stars} stars</span>
                  <span>{repo.forks} forks</span>
                  {repo.is_fork ? (
                    <Badge variant="secondary" className="text-[10px]">
                      Fork
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="text-[10px]">
                      Original
                    </Badge>
                  )}
                  {repo.language && (
                    <Badge variant="outline" className="text-[10px]">
                      {repo.language}
                    </Badge>
                  )}
                </div>
                {repo.pushed_at && (
                  <p className="mt-2 text-[11px] text-muted-foreground">
                    Updated {formatDistanceToNow(new Date(repo.pushed_at), { addSuffix: true })}
                  </p>
                )}
              </a>
              {repo.owner && (
                <p className="mt-2 text-[11px] text-muted-foreground">
                  by{" "}
                  <Link href={`/developers/${repo.owner}`} className="text-primary hover:underline">
                    @{repo.owner}
                  </Link>
                </p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
