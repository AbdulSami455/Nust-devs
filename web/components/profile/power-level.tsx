"use client";

import type { Developer } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export function powerTitle(level: number) {
  if (level >= 41) return "Legend";
  if (level >= 26) return "Maintainer";
  if (level >= 11) return "Builder";
  return "Contributor";
}

export function PowerLevelBadge({ dev }: { dev: Developer }) {
  const level = dev.power_level ?? 1;
  const title = powerTitle(level);
  const xp = dev.xp ?? 0;

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Badge variant="secondary" className="tabular-nums">
        Lv.{level} {title}
      </Badge>
      <span className="text-xs text-muted-foreground tabular-nums">{xp.toLocaleString()} XP</span>
      {(dev.current_streak ?? 0) > 0 && (
        <Badge
          variant="outline"
          className={cn(
            "tabular-nums",
            (dev.current_streak ?? 0) >= 7 && "border-orange-500/50 text-orange-600 dark:text-orange-400"
          )}
        >
          {(dev.streak_multiplier ?? 1) > 1
            ? `${dev.current_streak}d streak · ${dev.streak_multiplier}x`
            : `${dev.current_streak}d streak`}
        </Badge>
      )}
    </div>
  );
}
