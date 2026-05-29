import type { Metadata } from "next";
import { NavHeader } from "@/components/nav-header";
import { StatsClient } from "./client";

export const metadata: Metadata = {
  title: "Platform Stats — NUST Devs",
  description: "Platform-wide analytics for NUST developers on GitHub.",
};

export default function StatsPage() {
  return (
    <div className="min-h-screen bg-muted/40">
      <NavHeader />
      <StatsClient />
    </div>
  );
}
