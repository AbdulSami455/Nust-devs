"use client";

import Link from "next/link";
import { formatDistanceToNow } from "date-fns";
import type { ActivityEvent } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

export function LiveActivityFeed({
  events,
  loading,
}: {
  events: ActivityEvent[];
  loading: boolean;
}) {
  return (
    <div className="bento-card flex flex-col">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
          Recent Activity
        </h2>
        <span className="text-xs text-muted-foreground">Open source updates</span>
      </div>

      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 rounded-lg" />
          ))}
        </div>
      ) : events.length === 0 ? (
        <p className="text-sm text-muted-foreground">No recent repository activity yet.</p>
      ) : (
        <ul className="divide-y divide-border">
          {events.map((event, i) => (
            <li key={`${event.occurred_at}-${event.repo}-${i}`} className="flex gap-3 py-3 first:pt-0 last:pb-0">
              <div className="mt-1 size-2 shrink-0 rounded-full bg-primary" />
              <div className="min-w-0 flex-1">
                <p className="text-sm leading-snug">
                  <Link
                    href={`/developers/${event.username}`}
                    className="font-medium hover:underline"
                  >
                    @{event.username}
                  </Link>{" "}
                  <span className="text-muted-foreground">pushed to</span>{" "}
                  {event.repo ? (
                    <a
                      href={`https://github.com/${event.repo}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="font-mono text-sm hover:underline"
                    >
                      {event.repo}
                    </a>
                  ) : (
                    <span className="font-mono text-sm">a repository</span>
                  )}
                </p>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  {formatDistanceToNow(new Date(event.occurred_at), { addSuffix: true })}
                </p>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
