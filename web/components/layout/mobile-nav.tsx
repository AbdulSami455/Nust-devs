"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const links = [
  { href: "/", label: "Home" },
  { href: "/innovation", label: "Graph" },
  { href: "/developers", label: "Devs" },
  { href: "/leaderboard", label: "Ranks" },
  { href: "/projects", label: "Projects" },
  { href: "/stats", label: "Stats" },
];

export function MobileNav() {
  const path = usePathname();

  return (
    <nav className="fixed inset-x-0 bottom-0 z-50 border-t bg-background/95 backdrop-blur-md md:hidden">
      <div className="mx-auto flex max-w-lg items-stretch justify-around px-2 pb-[env(safe-area-inset-bottom)]">
        {links.map(({ href, label }) => {
          const active = path === href || (href !== "/" && path.startsWith(href));
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex flex-1 flex-col items-center gap-0.5 py-2.5 text-[10px] font-medium transition-colors",
                active ? "text-primary" : "text-muted-foreground"
              )}
            >
              <span
                className={cn(
                  "size-1 rounded-full",
                  active ? "bg-primary" : "bg-transparent"
                )}
              />
              {label}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
