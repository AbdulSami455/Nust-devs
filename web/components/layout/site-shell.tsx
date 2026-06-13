"use client";

import { usePathname } from "next/navigation";
import { NavHeader } from "@/components/nav-header";
import { MobileNav } from "@/components/layout/mobile-nav";
import { ChatWidget } from "@/components/ai/chat-widget";

export function SiteShell({ children }: { children: React.ReactNode }) {
  const path = usePathname();
  const isAdmin = path.startsWith("/admin");

  if (isAdmin) {
    return <>{children}</>;
  }

  return (
    <div className="min-h-screen bg-background pb-16 md:pb-0">
      <NavHeader />
      <main>{children}</main>
      <MobileNav />
      <ChatWidget />
    </div>
  );
}
