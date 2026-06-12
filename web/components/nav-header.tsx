"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { CommandMenu } from "@/components/command-menu";
import { ThemeToggle } from "@/components/theme-toggle";
import { cn } from "@/lib/utils";

const links = [
  { href: "/", label: "Home" },
  { href: "/join", label: "Join" },
  { href: "/innovation", label: "Innovation" },
  { href: "/developers", label: "Developers" },
  { href: "/leaderboard", label: "Leaderboard" },
  { href: "/hall-of-fame", label: "Hall of Fame" },
  { href: "/projects", label: "Projects" },
  { href: "/stats", label: "Stats" },
];

export function NavHeader() {
  const path = usePathname();

  return (
    <header className="glass-nav sticky top-0 z-50">
      <div className="mx-auto flex h-14 max-w-7xl items-center justify-between gap-4 px-4 sm:px-6">
        <Link href="/" className="flex items-center gap-2 shrink-0">
          <Image
            src="/nust-logo.svg"
            alt="NUST"
            width={34}
            height={34}
            priority
            className="size-8 rounded-full ring-1 ring-border/70"
          />
          <span className="hidden font-semibold tracking-tight sm:inline">NUST Devs</span>
        </Link>

        <nav className="hidden items-center gap-0.5 md:flex">
          {links.map(({ href, label }) => (
            <Link
              key={href}
              href={href}
              className={cn(
                "rounded-lg px-3 py-1.5 text-sm transition-colors",
                path === href
                  ? "bg-secondary font-medium text-foreground"
                  : "text-muted-foreground hover:bg-secondary hover:text-foreground"
              )}
            >
              {label}
            </Link>
          ))}
        </nav>

        <div className="flex items-center gap-1 sm:gap-2">
          <CommandMenu />
          <ThemeToggle />
        </div>
      </div>
    </header>
  );
}
