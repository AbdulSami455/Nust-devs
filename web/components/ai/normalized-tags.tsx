"use client";

import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";
import {
  fetchDeveloperNormalizedTags,
  fetchProjectNormalizedTags,
  type NormalizedTags,
} from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

type Props =
  | { kind: "developer"; username: string }
  | { kind: "project"; repoId: string };

export function NormalizedTagsCard(props: Props) {
  const [data, setData] = useState<NormalizedTags | null>(null);
  const [loading, setLoading] = useState(true);
  const isDeveloper = props.kind === "developer";
  const entityKey = isDeveloper ? props.username : props.repoId;

  useEffect(() => {
    let cancelled = false;
    setLoading(true);

    const request = isDeveloper
      ? fetchDeveloperNormalizedTags(entityKey)
      : fetchProjectNormalizedTags(entityKey);

    request
      .then((value) => {
        if (!cancelled) setData(value);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [isDeveloper, entityKey]);

  if (loading) {
    return (
      <div className="bento-card space-y-3">
        <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <Sparkles className="size-3.5" />
          AI normalized tags
        </div>
        <Skeleton className="h-4 w-44" />
        <Skeleton className="h-3 w-full" />
        <div className="flex flex-wrap gap-2">
          <Skeleton className="h-5 w-20 rounded-full" />
          <Skeleton className="h-5 w-24 rounded-full" />
          <Skeleton className="h-5 w-28 rounded-full" />
        </div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="bento-card space-y-2">
        <p className="text-[11px] uppercase tracking-wider text-muted-foreground">
          AI normalized tags
        </p>
        <p className="text-sm text-muted-foreground">Normalized tags unavailable.</p>
      </div>
    );
  }

  return (
    <section className="bento-card space-y-3 border-primary/15 bg-gradient-to-br from-primary/5 via-card to-card">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Sparkles className="size-3.5" />
        AI normalized tags
      </div>
      <p className="text-sm font-medium leading-snug">{data.headline}</p>
      <p className="text-xs leading-relaxed text-muted-foreground">{data.summary}</p>
      <TagGroup label="Languages" items={data.languages} />
      <TagGroup label="Skills" items={data.skills} />
      <TagGroup label="Tags" items={data.tags} />
    </section>
  );
}

function TagGroup({ label, items }: { label: string; items: string[] }) {
  if (items.length === 0) return null;
  return (
    <div className="space-y-2">
      <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
        {label}
      </p>
      <div className="flex flex-wrap gap-2">
        {items.map((item) => (
          <span
            key={item}
            className="inline-flex items-center rounded-full border border-border bg-background/70 px-2.5 py-0.5 text-xs text-foreground"
          >
            {item}
          </span>
        ))}
      </div>
    </div>
  );
}
