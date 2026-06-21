import type { Metadata } from "next";
import { Suspense } from "react";
import { CompareClient } from "./client";

export const metadata: Metadata = {
  title: "Compare Developers — NUST Devs",
  description: "Compare two tracked NUST developers side by side with AI-written takeaways.",
};

export default function ComparePage() {
  return (
    <Suspense
      fallback={
        <div className="mx-auto max-w-6xl space-y-8 px-4 py-8 sm:px-6">
          <div className="space-y-2">
            <div className="h-8 w-64 rounded-lg bg-muted" />
            <div className="h-4 w-96 rounded-lg bg-muted" />
          </div>
          <div className="h-40 rounded-2xl bg-muted/60" />
          <div className="h-80 rounded-2xl bg-muted/60" />
        </div>
      }
    >
      <CompareClient />
    </Suspense>
  );
}
