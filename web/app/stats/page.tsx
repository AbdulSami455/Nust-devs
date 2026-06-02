import type { Metadata } from "next";
import { StatsClient } from "./client";

export const metadata: Metadata = {
  title: "Platform Stats — NUST Devs",
  description: "Platform-wide analytics for NUST developers on GitHub.",
};

export default function StatsPage() {
  return <StatsClient />;
}
