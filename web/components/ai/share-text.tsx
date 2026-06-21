"use client";

import { useEffect, useState } from "react";
import { Sparkles, Copy } from "lucide-react";
import {
  fetchDeveloperShareText,
  fetchProjectShareText,
  type ShareTextInsight,
} from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";

type ShareTextCardProps =
  | { kind: "developer"; username: string; href?: string }
  | { kind: "project"; repoId: string; href: string };

export function ShareTextCard(props: ShareTextCardProps) {
  const [insight, setInsight] = useState<ShareTextInsight | null>(null);
  const [loading, setLoading] = useState(true);
  const entityKey = props.kind === "developer" ? props.username : props.repoId;

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setInsight(null);
    const loader =
      props.kind === "developer"
        ? fetchDeveloperShareText(props.username)
        : fetchProjectShareText(props.repoId);

    loader
      .then((data) => {
        if (!cancelled) setInsight(data);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [props.kind, props.username, props.repoId, entityKey]);

  const copy = async () => {
    if (!insight) return;
    const url = props.href ?? (typeof window !== "undefined" ? window.location.href : "");
    const text = url ? `${insight.share_text} ${url}` : insight.share_text;
    try {
      await navigator.clipboard.writeText(text);
      toast.success("Share text copied");
    } catch {
      toast.error("Could not copy share text");
    }
  };

  if (loading) {
    return (
      <section className="mt-3 space-y-2 rounded-xl border border-primary/10 bg-primary/5 p-3">
        <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <Sparkles className="size-3.5" />
          AI share text
        </div>
        <Skeleton className="h-3 w-40" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-8 w-28 rounded-lg" />
      </section>
    );
  }

  if (!insight) {
    return null;
  }

  return (
    <section className="mt-3 rounded-xl border border-primary/10 bg-primary/5 p-3">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Sparkles className="size-3.5" />
        AI share text
      </div>
      <p className="mt-2 text-sm font-medium leading-snug">{insight.headline}</p>
      <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{insight.share_text}</p>
      <div className="mt-3 flex gap-2">
        <Button size="sm" variant="outline" onClick={copy}>
          <Copy className="mr-2 size-3.5" />
          Copy share text
        </Button>
      </div>
    </section>
  );
}
