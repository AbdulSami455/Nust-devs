import type { Metadata } from "next";
import { ProjectsClient } from "./client";

export const metadata: Metadata = {
  title: "Top Projects — NUST Devs",
  description: "Top open-source repositories from NUST developers on GitHub.",
};

export default function ProjectsPage() {
  return <ProjectsClient />;
}
