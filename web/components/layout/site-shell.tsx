"use client";

import { usePathname } from "next/navigation";
import { NavHeader } from "@/components/nav-header";

export function SiteShell({ children }: { children: React.ReactNode }) {
  const path = usePathname();
  const isAdmin = path.startsWith("/admin");

  if (isAdmin) {
    return <>{children}</>;
  }

  return (
    <div className="min-h-screen bg-background">
      <NavHeader />
      <main>{children}</main>
    </div>
  );
}
