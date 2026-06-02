import type { Metadata } from "next";
import { InnovationClient } from "./client";

export const metadata: Metadata = {
  title: "Innovation Graph — NUST Devs",
  description: "Quarterly trends and live activity from NUST developers on GitHub.",
};

export default function InnovationPage() {
  return <InnovationClient />;
}
