"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Command } from "cmdk";
import {
  Code2,
  Home,
  Search,
  Trophy,
  Users,
  FolderGit2,
  BarChart3,
  Settings,
} from "lucide-react";
import { api, type Developer } from "@/lib/api";
import { Dialog, DialogContent } from "@/components/ui/dialog";

const pages = [
  { href: "/", label: "Home", icon: Home },
  { href: "/developers", label: "Developers", icon: Users },
  { href: "/leaderboard", label: "Leaderboard", icon: Trophy },
  { href: "/projects", label: "Projects", icon: FolderGit2 },
  { href: "/stats", label: "Stats", icon: BarChart3 },
  { href: "/admin", label: "Admin", icon: Settings },
];

export function CommandMenu() {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [developers, setDevelopers] = useState<Developer[]>([]);

  useEffect(() => {
    api.public.developers.list(1, 100).then(setDevelopers).catch(() => {});
  }, []);

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((v) => !v);
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, []);

  const go = useCallback(
    (href: string) => {
      setOpen(false);
      setQuery("");
      router.push(href);
    },
    [router]
  );

  const q = query.toLowerCase();
  const filteredDevs = developers.filter(
    (d) =>
      d.github_username.toLowerCase().includes(q) ||
      (d.display_name ?? "").toLowerCase().includes(q)
  );

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="hidden sm:flex items-center gap-2 rounded-lg border border-border/60 bg-muted/40 px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted/70 hover:text-foreground"
      >
        <Search className="size-4" />
        <span>Search…</span>
        <kbd className="ml-6 rounded border border-border bg-background px-1.5 py-0.5 text-[10px] font-medium">
          ⌘K
        </kbd>
      </button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="overflow-hidden p-0 sm:max-w-lg">
          <Command className="[&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:text-muted-foreground [&_[cmdk-group]:not([hidden])_~[cmdk-group]]:pt-0 [&_[cmdk-group]]:px-2 [&_[cmdk-input-wrapper]_svg]:size-5 [&_[cmdk-input]]:h-12 [&_[cmdk-item]]:px-2 [&_[cmdk-item]]:py-3 [&_[cmdk-item]_svg]:size-4">
            <div className="flex items-center border-b px-3">
              <Search className="mr-2 size-4 shrink-0 opacity-50" />
              <Command.Input
                placeholder="Search developers, pages…"
                value={query}
                onValueChange={setQuery}
                className="flex h-12 w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
              />
            </div>
            <Command.List className="max-h-80 overflow-y-auto p-2">
              <Command.Empty className="py-6 text-center text-sm text-muted-foreground">
                No results found.
              </Command.Empty>

              <Command.Group heading="Pages">
                {pages
                  .filter((p) => p.label.toLowerCase().includes(q))
                  .map(({ href, label, icon: Icon }) => (
                    <Command.Item
                      key={href}
                      value={label}
                      onSelect={() => go(href)}
                      className="flex cursor-pointer items-center gap-2 rounded-lg px-2 py-2 text-sm aria-selected:bg-accent"
                    >
                      <Icon className="size-4 text-muted-foreground" />
                      {label}
                    </Command.Item>
                  ))}
              </Command.Group>

              {filteredDevs.length > 0 && (
                <Command.Group heading="Developers">
                  {filteredDevs.slice(0, 8).map((dev) => (
                    <Command.Item
                      key={dev.id}
                      value={`${dev.github_username} ${dev.display_name ?? ""}`}
                      onSelect={() => go(`/developers/${dev.github_username}`)}
                      className="flex cursor-pointer items-center gap-3 rounded-lg px-2 py-2 text-sm aria-selected:bg-accent"
                    >
                      {dev.avatar_url ? (
                        <img src={dev.avatar_url} alt="" className="size-7 rounded-full" />
                      ) : (
                        <div className="flex size-7 items-center justify-center rounded-full bg-muted text-xs font-bold">
                          {dev.github_username[0]?.toUpperCase()}
                        </div>
                      )}
                      <div className="min-w-0">
                        <p className="truncate font-medium">{dev.display_name ?? dev.github_username}</p>
                        <p className="truncate text-xs text-muted-foreground">@{dev.github_username}</p>
                      </div>
                    </Command.Item>
                  ))}
                </Command.Group>
              )}

              <Command.Group heading="Actions">
                <Command.Item
                  onSelect={() => go("/developers")}
                  className="flex cursor-pointer items-center gap-2 rounded-lg px-2 py-2 text-sm aria-selected:bg-accent"
                >
                  <Code2 className="size-4 text-muted-foreground" />
                  Browse all developers
                </Command.Item>
              </Command.Group>
            </Command.List>
          </Command>
        </DialogContent>
      </Dialog>
    </>
  );
}
