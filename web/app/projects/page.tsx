import type { Metadata } from "next";
import { NavHeader } from "@/components/nav-header";
import { ProjectsClient } from "./client";

export const metadata: Metadata = {
  title: "Top Projects — NUST Devs",
  description: "Top open-source repositories from NUST developers on GitHub.",
};

export default function ProjectsPage() {
  return (
    <div className="min-h-screen bg-muted/40">
      <NavHeader />
      <ProjectsClient />
    </div>
  );
}
