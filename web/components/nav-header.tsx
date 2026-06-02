"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { CommandMenu } from "@/components/command-menu";
import { ThemeToggle } from "@/components/theme-toggle";
import { cn } from "@/lib/utils";

const links = [
  { href: "/", label: "Home" },
  { href: "/developers", label: "Developers" },
  { href: "/leaderboard", label: "Leaderboard" },
  { href: "/projects", label: "Projects" },
  { href: "/stats", label: "Stats" },
];

export function NavHeader() {
  const path = usePathname();

  return (
    <header className="glass-nav sticky top-0 z-50">
      <div className="mx-auto flex h-14 max-w-7xl items-center justify-between gap-4 px-4 sm:px-6">
        <Link href="/" className="flex items-center gap-2 shrink-0">
          <div className="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground text-xs font-bold">
            N
          </div>
          <span className="hidden font-bold tracking-tight sm:inline">NUST Devs</span>
        </Link>

        <nav className="hidden items-center gap-0.5 md:flex">
          {links.map(({ href, label }) => (
            <Link
              key={href}
              href={href}
              className={cn(
                "rounded-lg px-3 py-1.5 text-sm transition-colors",
                path === href
                  ? "bg-primary/15 font-medium text-primary"
                  : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"
              )}
            >
              {label}
            </Link>
          ))}
        </nav>

        <div className="flex items-center gap-1 sm:gap-2">
          <CommandMenu />
          <ThemeToggle />
          <Link
            href="/admin"
            className="rounded-lg px-2.5 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted/60 hover:text-foreground"
          >
            Admin
          </Link>
        </div>
      </div>
    </header>
  );
}
