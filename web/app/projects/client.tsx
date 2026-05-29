"use client";

import { useEffect, useState } from "react";
import { api, type PublicRepo } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";

const LANG_COLORS: Record<string, string> = {
  Go:         "bg-sky-400",
  TypeScript: "bg-blue-500",
  JavaScript: "bg-yellow-400",
  Python:     "bg-green-500",
  Rust:       "bg-orange-500",
  Java:       "bg-red-500",
  "C++":      "bg-pink-500",
  C:          "bg-gray-500",
};

export function ProjectsClient() {
  const [repos, setRepos] = useState<PublicRepo[]>([]);
  const [query, setQuery] = useState("");
  const [lang, setLang] = useState("All");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.public.topProjects().then(setRepos).catch(() => {}).finally(() => setLoading(false));
  }, []);

  const languages = ["All", ...Array.from(new Set(repos.map((r) => r.language).filter(Boolean) as string[]))];

  const filtered = repos.filter((r) => {
    const matchQ = r.full_name.toLowerCase().includes(query.toLowerCase()) ||
                   r.description?.toLowerCase().includes(query.toLowerCase());
    const matchL = lang === "All" || r.language === lang;
    return matchQ && matchL;
  });

  return (
    <main className="mx-auto max-w-6xl px-6 py-8 space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <h2 className="text-2xl font-bold">Top Projects</h2>
        <Input
          placeholder="Search repos…"
          className="max-w-xs"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
      </div>

      {/* Language filter */}
      <div className="flex flex-wrap gap-2">
        {languages.map((l) => (
          <button
            key={l}
            onClick={() => setLang(l)}
            className={`px-3 py-1 rounded-full text-xs font-medium transition-colors ${
              lang === l ? "bg-primary text-primary-foreground" : "bg-muted hover:bg-muted/80"
            }`}
          >
            {l}
          </button>
        ))}
      </div>

      {loading ? (
        <p className="text-muted-foreground">Loading…</p>
      ) : filtered.length === 0 ? (
        <p className="text-muted-foreground">No repositories found.</p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {filtered.map((repo) => (
            <a key={repo.id} href={repo.url} target="_blank" rel="noopener noreferrer">
              <Card className="hover:border-foreground/30 transition-colors h-full cursor-pointer">
                <CardContent className="pt-4 flex flex-col gap-2 h-full">
                  <div className="flex items-start justify-between gap-2">
                    <p className="font-mono font-medium text-sm leading-snug">{repo.full_name}</p>
                    {repo.language && (
                      <span className={`w-2.5 h-2.5 rounded-full shrink-0 mt-1 ${LANG_COLORS[repo.language] ?? "bg-muted-foreground"}`} />
                    )}
                  </div>
                  <p className="text-xs text-muted-foreground line-clamp-2 flex-1">
                    {repo.description || "No description"}
                  </p>
                  <div className="flex items-center gap-3 pt-1 text-xs text-muted-foreground">
                    <span className="flex items-center gap-1">★ {repo.stars}</span>
                    <span className="flex items-center gap-1">⑂ {repo.forks}</span>
                    {repo.language && <Badge variant="outline" className="text-xs">{repo.language}</Badge>}
                  </div>
                </CardContent>
              </Card>
            </a>
          ))}
        </div>
      )}
    </main>
  );
}
