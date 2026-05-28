"use client";

import { useEffect, useState, use } from "react";
import Link from "next/link";
import { api, type Developer, type PublicRepo, type ContributionDay } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function DeveloperProfilePage({ params }: { params: Promise<{ username: string }> }) {
  const { username } = use(params);
  const [dev, setDev] = useState<Developer | null>(null);
  const [repos, setRepos] = useState<PublicRepo[]>([]);
  const [contributions, setContributions] = useState<ContributionDay[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      api.public.developers.get(username),
      api.public.developers.repos(username),
      api.public.developers.contributions(username),
    ]).then(([d, r, c]) => {
      setDev(d);
      setRepos(r ?? []);
      setContributions(c ?? []);
    }).catch(() => {}).finally(() => setLoading(false));
  }, [username]);

  if (loading) return <div className="min-h-screen flex items-center justify-center text-muted-foreground">Loading…</div>;
  if (!dev) return <div className="min-h-screen flex items-center justify-center text-muted-foreground">Developer not found.</div>;

  const maxCount = Math.max(...contributions.map((d) => d.count), 1);

  return (
    <div className="min-h-screen bg-muted/40">
      <header className="border-b bg-background px-6 py-4 flex items-center justify-between">
        <Link href="/" className="text-xl font-bold">NUST Devs</Link>
        <nav className="flex gap-4 text-sm">
          <Link href="/developers" className="hover:underline">Developers</Link>
          <Link href="/leaderboard" className="hover:underline">Leaderboard</Link>
        </nav>
      </header>

      <main className="mx-auto max-w-5xl px-6 py-8 space-y-8">
        {/* Profile header */}
        <div className="flex gap-6 items-start">
          {dev.avatar_url ? (
            <img src={dev.avatar_url} alt={dev.github_username} className="w-24 h-24 rounded-full border" />
          ) : (
            <div className="w-24 h-24 rounded-full bg-muted flex items-center justify-center text-3xl font-bold text-muted-foreground">
              {dev.github_username[0].toUpperCase()}
            </div>
          )}
          <div className="space-y-1">
            <div className="flex items-center gap-3">
              <h2 className="text-2xl font-bold">{dev.display_name ?? dev.github_username}</h2>
              {dev.verification_status === "email_verified" && <Badge>verified</Badge>}
            </div>
            <p className="text-muted-foreground font-mono">@{dev.github_username}</p>
            {dev.bio && <p className="text-sm max-w-lg">{dev.bio}</p>}
            <div className="flex gap-4 text-sm text-muted-foreground">
              {dev.location && <span>📍 {dev.location}</span>}
              {dev.company && <span>🏢 {dev.company}</span>}
              {dev.website && <a href={dev.website} target="_blank" rel="noopener noreferrer" className="hover:underline">🔗 {dev.website}</a>}
            </div>
            <a href={`https://github.com/${dev.github_username}`} target="_blank" rel="noopener noreferrer"
               className="text-sm text-muted-foreground hover:underline inline-block">
              View on GitHub →
            </a>
          </div>
        </div>

        {/* Stat cards */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          {[
            { label: "Stars", value: dev.total_stars },
            { label: "Repos", value: dev.public_repos },
            { label: "Followers", value: dev.followers },
            { label: "Activity Score", value: Math.round(dev.activity_score) },
          ].map(({ label, value }) => (
            <Card key={label}>
              <CardHeader className="pb-1">
                <CardTitle className="text-sm font-medium text-muted-foreground">{label}</CardTitle>
              </CardHeader>
              <CardContent><p className="text-2xl font-bold">{value}</p></CardContent>
            </Card>
          ))}
        </div>

        {/* Contribution heatmap */}
        {contributions.length > 0 && (
          <section className="space-y-2">
            <h3 className="font-semibold">Contributions (last year)</h3>
            <div className="flex flex-wrap gap-0.5">
              {contributions.map((d) => {
                const intensity = d.count === 0 ? 0 : Math.ceil((d.count / maxCount) * 4);
                const colors = ["bg-muted", "bg-green-200", "bg-green-400", "bg-green-600", "bg-green-800"];
                return (
                  <div key={d.date} title={`${d.date}: ${d.count}`}
                       className={`w-2.5 h-2.5 rounded-sm ${colors[intensity]}`} />
                );
              })}
            </div>
            <p className="text-xs text-muted-foreground">
              {contributions.reduce((s, d) => s + d.count, 0)} contributions in the last year
            </p>
          </section>
        )}

        {/* Repositories */}
        <section className="space-y-3">
          <h3 className="font-semibold">Repositories ({repos.length})</h3>
          <div className="rounded-lg border bg-background overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Repository</TableHead>
                  <TableHead>Language</TableHead>
                  <TableHead>Stars</TableHead>
                  <TableHead>Forks</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {repos.slice(0, 20).map((repo) => (
                  <TableRow key={repo.id}>
                    <TableCell>
                      <a href={repo.url} target="_blank" rel="noopener noreferrer" className="font-mono hover:underline font-medium">
                        {repo.name}
                      </a>
                      {repo.is_fork && <Badge variant="secondary" className="ml-2 text-xs">fork</Badge>}
                      {repo.description && <p className="text-xs text-muted-foreground mt-0.5 truncate max-w-xs">{repo.description}</p>}
                    </TableCell>
                    <TableCell>{repo.language ? <Badge variant="outline">{repo.language}</Badge> : "—"}</TableCell>
                    <TableCell>{repo.stars}</TableCell>
                    <TableCell>{repo.forks}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </section>
      </main>
    </div>
  );
}
