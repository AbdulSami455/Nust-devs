"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const links = [
  { href: "/",            label: "Home"        },
  { href: "/developers",  label: "Developers"  },
  { href: "/leaderboard", label: "Leaderboard" },
  { href: "/projects",    label: "Projects"    },
  { href: "/stats",       label: "Stats"       },
];

export function NavHeader() {
  const path = usePathname();
  return (
    <header className="border-b bg-background px-6 py-4 flex items-center justify-between">
      <Link href="/" className="text-xl font-bold tracking-tight">NUST Devs</Link>
      <nav className="flex items-center gap-1">
        {links.map(({ href, label }) => (
          <Link
            key={href}
            href={href}
            className={`px-3 py-1.5 rounded-md text-sm transition-colors ${
              path === href
                ? "bg-muted font-medium"
                : "text-muted-foreground hover:text-foreground hover:bg-muted/50"
            }`}
          >
            {label}
          </Link>
        ))}
        <Link
          href="/admin"
          className="ml-2 px-3 py-1.5 rounded-md text-sm text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors"
        >
          Admin
        </Link>
      </nav>
    </header>
  );
}
