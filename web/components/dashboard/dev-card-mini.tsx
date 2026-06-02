"use client";

import Link from "next/link";
import type { Developer } from "@/lib/api";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";

export function DevCardMini({ dev }: { dev: Developer }) {
  return (
    <Link href={`/developers/${dev.github_username}`} className="group block">
      <div className="bento-card h-full transition-all hover:border-primary/40 hover:shadow-md hover:shadow-primary/5">
        <div className="flex items-start gap-3">
          <Avatar className="size-11">
            {dev.avatar_url && <AvatarImage src={dev.avatar_url} alt={dev.github_username} />}
            <AvatarFallback>{dev.github_username[0]?.toUpperCase()}</AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1">
            <p className="truncate font-semibold group-hover:text-primary transition-colors">
              {dev.display_name ?? dev.github_username}
            </p>
            <p className="truncate font-mono text-xs text-muted-foreground">@{dev.github_username}</p>
          </div>
          {dev.verification_status === "email_verified" && (
            <Badge className="shrink-0 text-[10px]">verified</Badge>
          )}
        </div>
        <div className="mt-4 flex gap-3 text-xs text-muted-foreground">
          <span>{dev.total_stars} stars</span>
          <span>{dev.public_repos} repos</span>
          <span>{dev.followers} followers</span>
        </div>
      </div>
    </Link>
  );
}
