"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { streamChat, type ChatAgentEvent, type ChatMessage } from "@/lib/api";
import { cn } from "@/lib/utils";

// ── Suggested prompts ─────────────────────────────────────────────────────────
const SUGGESTED: { icon: string; label: string; prompt: string }[] = [
  { icon: "🏆", label: "Top developers",    prompt: "Who are the top 10 developers right now?" },
  { icon: "🔍", label: "Find by language",  prompt: "Find Go developers with 50+ stars" },
  { icon: "🔥", label: "Longest streaks",   prompt: "Who has the longest contribution streak?" },
  { icon: "📈", label: "Trending projects", prompt: "What are the fastest growing NUST projects?" },
  { icon: "🤝", label: "Best reviewers",    prompt: "Show me developers with the most code reviews" },
  { icon: "🌐", label: "Community stats",   prompt: "Give me an overview of the NUST dev community" },
];

// ── Types ─────────────────────────────────────────────────────────────────────
interface Message {
  role: "user" | "assistant";
  content: string;
  streaming?: boolean;
  ts: number;
}

interface AgentTimelineEvent {
  id: string;
  type: ChatAgentEvent["type"];
  message: string;
  toolName?: string;
  success?: boolean;
  latencyMS?: number;
}

// ── Language colour map ───────────────────────────────────────────────────────
const LANG_COLOR: Record<string, string> = {
  python:     "bg-blue-400/20 text-blue-600 dark:text-blue-300",
  javascript: "bg-yellow-400/20 text-yellow-600 dark:text-yellow-300",
  typescript: "bg-sky-400/20 text-sky-600 dark:text-sky-300",
  go:         "bg-cyan-400/20 text-cyan-600 dark:text-cyan-300",
  rust:       "bg-orange-400/20 text-orange-600 dark:text-orange-300",
  java:       "bg-red-400/20 text-red-600 dark:text-red-300",
  "c++":      "bg-purple-400/20 text-purple-600 dark:text-purple-300",
  c:          "bg-indigo-400/20 text-indigo-600 dark:text-indigo-300",
  kotlin:     "bg-violet-400/20 text-violet-600 dark:text-violet-300",
  swift:      "bg-orange-400/20 text-orange-500 dark:text-orange-300",
  ruby:       "bg-red-400/20 text-red-500 dark:text-red-300",
  php:        "bg-indigo-400/20 text-indigo-500 dark:text-indigo-300",
  shell:      "bg-green-400/20 text-green-600 dark:text-green-300",
  html:       "bg-orange-300/20 text-orange-500 dark:text-orange-300",
  css:        "bg-blue-300/20 text-blue-500 dark:text-blue-300",
};
function langClass(lang?: string) {
  if (!lang) return "bg-muted text-muted-foreground";
  return LANG_COLOR[lang.toLowerCase()] ?? "bg-muted text-muted-foreground";
}

// ── Structured-content parser ─────────────────────────────────────────────────
// Detects blocks like markdown tables, repo/dev bullet patterns, and stat lines.

interface ParsedBlock {
  type: "text" | "table" | "project-cards" | "dev-cards" | "stat-grid";
  raw: string;
  data?: unknown;
}

interface TableData { headers: string[]; rows: string[][] }
interface ProjectCard { rank?: number; name: string; owner?: string; language?: string; stars: number; forks?: number; description?: string; url?: string }
interface DevCard { rank?: number; username: string; language?: string; score?: string; streak?: number; stars?: number; prs?: number; reviews?: number; powerTitle?: string }
interface StatItem { label: string; value: string; sub?: string }

function parseMarkdownTable(block: string): TableData | null {
  const lines = block.trim().split("\n").filter((l) => l.includes("|"));
  if (lines.length < 3) return null;
  const headers = lines[0].split("|").map((h) => h.trim()).filter(Boolean);
  const rows = lines.slice(2).map((l) => l.split("|").map((c) => c.trim()).filter(Boolean));
  if (headers.length === 0 || rows.length === 0) return null;
  return { headers, rows };
}

/** Try to extract project cards from a markdown table row. */
function tableToProjects(td: TableData): ProjectCard[] | null {
  const lower = td.headers.map((h) => h.toLowerCase());
  const projIdx = lower.findIndex((h) => h.includes("project") || h.includes("repo") || h.includes("name"));
  const ownerIdx = lower.findIndex((h) => h.includes("owner") || h.includes("user"));
  const langIdx  = lower.findIndex((h) => h.includes("lang"));
  const starsIdx = lower.findIndex((h) => h.includes("star"));
  const forksIdx = lower.findIndex((h) => h.includes("fork"));
  const descIdx  = lower.findIndex((h) => h.includes("desc"));

  if (projIdx === -1 || starsIdx === -1) return null;

  return td.rows.map((row, i) => ({
    rank:        i + 1,
    name:        stripMd(row[projIdx] ?? ""),
    owner:       ownerIdx >= 0 ? stripMd(row[ownerIdx] ?? "") : undefined,
    language:    langIdx >= 0  ? stripMd(row[langIdx]  ?? "") : undefined,
    stars:       parseInt(row[starsIdx]?.replace(/[^0-9]/g, "") ?? "0") || 0,
    forks:       forksIdx >= 0 ? parseInt(row[forksIdx]?.replace(/[^0-9]/g, "") ?? "0") || 0 : undefined,
    description: descIdx >= 0  ? stripMd(row[descIdx]  ?? "") : undefined,
  }));
}

/** Try to extract dev cards from a markdown table. */
function tableToDevelopers(td: TableData): DevCard[] | null {
  const lower = td.headers.map((h) => h.toLowerCase());
  const nameIdx    = lower.findIndex((h) => h.includes("developer") || h.includes("dev") || h.includes("name") || h.includes("user"));
  const scoreIdx   = lower.findIndex((h) => h.includes("score"));
  const langIdx    = lower.findIndex((h) => h.includes("lang"));
  const starsIdx   = lower.findIndex((h) => h.includes("star"));
  const streakIdx  = lower.findIndex((h) => h.includes("streak"));
  const prsIdx     = lower.findIndex((h) => h.includes("pr") || h.includes("pull"));
  const reviewIdx  = lower.findIndex((h) => h.includes("review"));
  const titleIdx   = lower.findIndex((h) => h.includes("title") || h.includes("level") || h.includes("power"));

  if (nameIdx === -1) return null;

  return td.rows.map((row, i) => ({
    rank:       i + 1,
    username:   stripMd(row[nameIdx] ?? ""),
    score:      scoreIdx  >= 0 ? stripMd(row[scoreIdx]  ?? "") : undefined,
    language:   langIdx   >= 0 ? stripMd(row[langIdx]   ?? "") : undefined,
    stars:      starsIdx  >= 0 ? parseInt(row[starsIdx]?.replace(/[^0-9]/g, "") ?? "0") || 0 : undefined,
    streak:     streakIdx >= 0 ? parseInt(row[streakIdx]?.replace(/[^0-9]/g, "") ?? "0") || 0 : undefined,
    prs:        prsIdx    >= 0 ? parseInt(row[prsIdx]?.replace(/[^0-9]/g, "") ?? "0") || 0 : undefined,
    reviews:    reviewIdx >= 0 ? parseInt(row[reviewIdx]?.replace(/[^0-9]/g, "") ?? "0") || 0 : undefined,
    powerTitle: titleIdx  >= 0 ? stripMd(row[titleIdx]  ?? "") : undefined,
  }));
}

/** Strip markdown bold/italic/link from a cell. */
function stripMd(s: string) {
  return s.replace(/\*\*(.+?)\*\*/g, "$1").replace(/\[(.+?)\]\(.+?\)/g, "$1").replace(/[*_]/g, "").trim();
}

/**
 * Split the assistant message into typed blocks.
 * Tables → try project/dev cards first, fall back to generic table.
 * Everything else → text block.
 */
function parseBlocks(text: string): ParsedBlock[] {
  const blocks: ParsedBlock[] = [];
  // Split on markdown table blocks (lines that start with |)
  const tableRe = /((?:\|.+\n?){3,})/g;
  let last = 0;
  let m: RegExpExecArray | null;

  while ((m = tableRe.exec(text)) !== null) {
    // text before table
    const pre = text.slice(last, m.index).trim();
    if (pre) blocks.push(...splitTextBlock(pre));

    const raw = m[1].trim();
    const td = parseMarkdownTable(raw);
    if (td) {
      const projects = tableToProjects(td);
      if (projects && projects.length > 0 && projects.some((p) => p.stars > 0)) {
        blocks.push({ type: "project-cards", raw, data: projects });
      } else {
        const devs = tableToDevelopers(td);
        if (devs && devs.length > 0) {
          blocks.push({ type: "dev-cards", raw, data: devs });
        } else {
          blocks.push({ type: "table", raw, data: td });
        }
      }
    } else {
      blocks.push({ type: "text", raw });
    }
    last = m.index + m[0].length;
  }

  const tail = text.slice(last).trim();
  if (tail) blocks.push(...splitTextBlock(tail));

  return blocks;
}

/** Look for "Label: value" stat lines in a text block and extract them. */
function splitTextBlock(text: string): ParsedBlock[] {
  return [{ type: "text", raw: text }];
}

// ── Rich-content renderers ─────────────────────────────────────────────────────

function ProjectCards({ cards }: { cards: ProjectCard[] }) {
  return (
    <div className="mt-2 space-y-2">
      {cards.map((p, i) => (
        <div
          key={i}
          className="rounded-xl border border-border/70 bg-background/80 p-3 ring-1 ring-black/[0.03] dark:ring-white/[0.04]"
        >
          <div className="flex items-start justify-between gap-2">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-1.5 flex-wrap">
                {p.rank && (
                  <span className="text-[10px] font-semibold text-muted-foreground tabular-nums">#{p.rank}</span>
                )}
                <span className="truncate font-mono text-sm font-semibold text-foreground">
                  {p.owner ? `${p.owner}/` : ""}
                  <span className="text-primary">{p.name}</span>
                </span>
              </div>
              {p.description && (
                <p className="mt-0.5 line-clamp-2 text-[11px] text-muted-foreground leading-relaxed">{p.description}</p>
              )}
            </div>
            {p.language && (
              <span className={cn("shrink-0 rounded-full px-2 py-0.5 text-[10px] font-medium", langClass(p.language))}>
                {p.language}
              </span>
            )}
          </div>
          <div className="mt-2 flex items-center gap-3 text-[11px] text-muted-foreground">
            <span className="flex items-center gap-1">
              <svg className="size-3 text-yellow-500" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
              <strong className="text-foreground">{p.stars.toLocaleString()}</strong>
            </span>
            {p.forks !== undefined && (
              <span className="flex items-center gap-1">
                <svg className="size-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><circle cx="12" cy="18" r="3"/><circle cx="6" cy="6" r="3"/><circle cx="18" cy="6" r="3"/><path d="M6 9v2a2 2 0 002 2h8a2 2 0 002-2V9"/><path d="M12 12v3"/></svg>
                {p.forks.toLocaleString()} forks
              </span>
            )}
            {p.url && (
              <a
                href={p.url}
                target="_blank"
                rel="noopener noreferrer"
                className="ml-auto flex items-center gap-0.5 font-medium text-primary hover:underline"
              >
                View
                <svg className="size-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/></svg>
              </a>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}

function DeveloperCards({ cards }: { cards: DevCard[] }) {
  return (
    <div className="mt-2 space-y-2">
      {cards.map((d, i) => (
        <div
          key={i}
          className="rounded-xl border border-border/70 bg-background/80 p-3 ring-1 ring-black/[0.03] dark:ring-white/[0.04]"
        >
          <div className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0">
              {/* Rank badge */}
              <span className={cn(
                "flex size-6 shrink-0 items-center justify-center rounded-full text-[10px] font-bold",
                i === 0 ? "bg-yellow-400/20 text-yellow-600 dark:text-yellow-300" :
                i === 1 ? "bg-zinc-300/30 text-zinc-500 dark:text-zinc-300" :
                i === 2 ? "bg-orange-400/20 text-orange-600 dark:text-orange-300" :
                "bg-muted text-muted-foreground",
              )}>
                {d.rank ?? i + 1}
              </span>
              <div className="min-w-0">
                <a
                  href={`/developers/${d.username}`}
                  className="block truncate font-mono text-sm font-semibold text-primary hover:underline"
                >
                  @{d.username}
                </a>
                {d.powerTitle && (
                  <span className="text-[10px] text-muted-foreground">{d.powerTitle}</span>
                )}
              </div>
            </div>
            {d.language && (
              <span className={cn("shrink-0 rounded-full px-2 py-0.5 text-[10px] font-medium", langClass(d.language))}>
                {d.language}
              </span>
            )}
          </div>

          <div className="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
            {d.score !== undefined && (
              <span className="flex items-center gap-1">
                <svg className="size-3 text-primary" viewBox="0 0 24 24" fill="currentColor"><path d="M13 2.05v2.02c3.95.49 7 3.85 7 7.93 0 3.21-1.81 6-4.72 7.72L13 17v5h5l-1.22-1.22C19.91 19.07 22 15.76 22 12c0-5.18-3.95-9.45-9-9.95zM11 2.05C5.95 2.55 2 6.82 2 12c0 3.76 2.09 7.07 5.22 8.78L6 22h5v-5l-2.28 2.72C7.81 18 6 15.21 6 12c0-4.08 3.05-7.44 7-7.93V2.05z"/></svg>
                {d.score}
              </span>
            )}
            {d.stars !== undefined && (
              <span className="flex items-center gap-1">
                <svg className="size-3 text-yellow-500" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
                {d.stars.toLocaleString()} stars
              </span>
            )}
            {d.streak !== undefined && d.streak > 0 && (
              <span className="flex items-center gap-1">
                <span>🔥</span>
                {d.streak}d streak
              </span>
            )}
            {d.prs !== undefined && d.prs > 0 && (
              <span>{d.prs} PRs</span>
            )}
            {d.reviews !== undefined && d.reviews > 0 && (
              <span>{d.reviews} reviews</span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}

function GenericTable({ data }: { data: TableData }) {
  return (
    <div className="mt-2 overflow-x-auto rounded-xl border border-border/70">
      <table className="w-full text-[11px]">
        <thead>
          <tr className="border-b border-border/70 bg-muted/40">
            {data.headers.map((h, i) => (
              <th key={i} className="px-3 py-2 text-left font-semibold text-muted-foreground whitespace-nowrap">
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.rows.map((row, ri) => (
            <tr key={ri} className="border-b border-border/40 last:border-0 even:bg-muted/20">
              {row.map((cell, ci) => (
                <td key={ci} className="px-3 py-2 text-foreground whitespace-nowrap">
                  {stripMd(cell)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

/** Plain text with markdown: bold, inline-code, bullets, headings */
function TextBlock({ text }: { text: string }) {
  if (!text.trim()) return null;
  const lines = text.split("\n");
  return (
    <div className="space-y-1">
      {lines.map((line, li) => {
        if (!line.trim()) return <div key={li} className="h-1" />;

        // Heading
        const hMatch = line.match(/^(#{1,3})\s+(.+)/);
        if (hMatch) {
          const level = hMatch[1].length;
          const Comp = level === 1 ? "h3" : level === 2 ? "h4" : "h5";
          return (
            <Comp key={li} className={cn("font-semibold", level === 1 ? "text-sm mt-1" : "text-xs mt-0.5")}>
              {renderInline(hMatch[2])}
            </Comp>
          );
        }

        // Bullet
        const isBullet = /^[-*•]\s/.test(line);
        const content = isBullet ? line.replace(/^[-*•]\s/, "") : line;

        return (
          <p key={li} className={cn("leading-relaxed text-sm", isBullet && "flex gap-1.5")}>
            {isBullet && <span className="mt-[3px] shrink-0 text-primary/70 text-xs">•</span>}
            <span>{renderInline(content)}</span>
          </p>
        );
      })}
    </div>
  );
}

function renderInline(text: string) {
  const parts = text.split(/(\*\*[^*]+\*\*|`[^`]+`|\[.+?\]\(.+?\))/g);
  return parts.map((part, i) => {
    if (part.startsWith("**") && part.endsWith("**"))
      return <strong key={i}>{part.slice(2, -2)}</strong>;
    if (part.startsWith("`") && part.endsWith("`"))
      return (
        <code key={i} className="rounded bg-black/10 px-1 py-0.5 font-mono text-[11px] dark:bg-white/10">
          {part.slice(1, -1)}
        </code>
      );
    const linkMatch = part.match(/^\[(.+?)\]\((.+?)\)$/);
    if (linkMatch)
      return (
        <a key={i} href={linkMatch[2]} target="_blank" rel="noopener noreferrer" className="text-primary underline underline-offset-2">
          {linkMatch[1]}
        </a>
      );
    return <span key={i}>{part}</span>;
  });
}

function RichMessage({ content }: { content: string }) {
  const blocks = parseBlocks(content);
  return (
    <div className="space-y-1.5">
      {blocks.map((block, i) => {
        if (block.type === "project-cards")
          return <ProjectCards key={i} cards={block.data as ProjectCard[]} />;
        if (block.type === "dev-cards")
          return <DeveloperCards key={i} cards={block.data as DevCard[]} />;
        if (block.type === "table")
          return <GenericTable key={i} data={block.data as TableData} />;
        return <TextBlock key={i} text={block.raw} />;
      })}
    </div>
  );
}

function formatTime(ts: number) {
  return new Date(ts).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

// ── Icons ─────────────────────────────────────────────────────────────────────
const IconChat = () => (
  <svg className="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.8}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
  </svg>
);
const IconX = () => (
  <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
  </svg>
);
const IconSend = () => (
  <svg className="size-4" viewBox="0 0 24 24" fill="currentColor">
    <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
  </svg>
);
const IconStop = () => (
  <svg className="size-4" viewBox="0 0 24 24" fill="currentColor">
    <rect x="6" y="6" width="12" height="12" rx="2" />
  </svg>
);
const IconTrash = () => (
  <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.8}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
  </svg>
);
const IconCopy = () => (
  <svg className="size-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
  </svg>
);
const IconCheck = () => (
  <svg className="size-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.5}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
  </svg>
);
const IconMinus = () => (
  <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M20 12H4" />
  </svg>
);

// ── Typing dots ───────────────────────────────────────────────────────────────
function TypingDots() {
  return (
    <span className="inline-flex items-center gap-[3px] py-0.5">
      {[0, 150, 300].map((delay) => (
        <span key={delay} className="size-1.5 rounded-full bg-current opacity-60 animate-bounce"
          style={{ animationDelay: `${delay}ms`, animationDuration: "1s" }} />
      ))}
    </span>
  );
}

// ── Copy button ───────────────────────────────────────────────────────────────
function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <button onClick={() => {
      navigator.clipboard.writeText(text).then(() => {
        setCopied(true);
        setTimeout(() => setCopied(false), 1800);
      });
    }} title="Copy" className="rounded p-1 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 hover:text-foreground">
      {copied ? <IconCheck /> : <IconCopy />}
    </button>
  );
}

// ── Main component ─────────────────────────────────────────────────────────────
export function ChatWidget() {
  const [open, setOpen]           = useState(false);
  const [minimized, setMinimized] = useState(false);
  const [messages, setMessages]   = useState<Message[]>([]);
  const [input, setInput]         = useState("");
  const [busy, setBusy]           = useState(false);
  const [error, setError]         = useState<string | null>(null);
  const [agentEvents, setAgentEvents] = useState<AgentTimelineEvent[]>([]);
  const abortRef  = useRef<AbortController | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef  = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (!minimized) bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, minimized]);

  useEffect(() => {
    if (open && !minimized) setTimeout(() => inputRef.current?.focus(), 120);
  }, [open, minimized]);

  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === "Escape" && open) setOpen(false); };
    window.addEventListener("keydown", h);
    return () => window.removeEventListener("keydown", h);
  }, [open]);

  useEffect(() => {
    const el = inputRef.current;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = Math.min(el.scrollHeight, 120) + "px";
  }, [input]);

  const send = useCallback((text: string) => {
    const msg = text.trim();
    if (!msg || busy) return;
    setError(null);
    setInput("");

    const userMsg: Message = { role: "user", content: msg, ts: Date.now() };
    setMessages((prev) => [...prev, userMsg]);
    setAgentEvents([{ id: `${Date.now()}-status`, type: "status", message: "Thinking" }]);

    const history: ChatMessage[] = messages.map((m) => ({ role: m.role, content: m.content }));
    setMessages((prev) => [...prev, { role: "assistant", content: "", streaming: true, ts: Date.now() }]);
    setBusy(true);
    setMinimized(false);

    abortRef.current = streamChat(msg, history,
      (token) => {
        setMessages((prev) => {
          const copy = [...prev];
          const last = copy[copy.length - 1];
          if (last?.role === "assistant") copy[copy.length - 1] = { ...last, content: last.content + token };
          return copy;
        });
      },
      (event) => {
        setAgentEvents((prev) => [
          ...prev,
          {
            id: `${Date.now()}-${prev.length}`,
            type: event.type,
            message: event.message ?? describeAgentEvent(event),
            toolName: event.tool_name,
            success: event.success,
            latencyMS: event.latency_ms,
          },
        ]);
      },
      () => {
        setMessages((prev) => {
          const copy = [...prev];
          const last = copy[copy.length - 1];
          if (last?.role === "assistant") copy[copy.length - 1] = { ...last, streaming: false };
          return copy;
        });
        setBusy(false);
      },
      (err) => {
        setMessages((prev) => prev.filter((m) => !m.streaming));
        setAgentEvents((prev) => [...prev, { id: `${Date.now()}-error`, type: "status", message: "Run failed" }]);
        setError(err);
        setBusy(false);
      },
    );
  }, [busy, messages]);

  const cancel = () => {
    abortRef.current?.abort();
    setMessages((prev) => prev.filter((m) => !m.streaming));
    setAgentEvents((prev) => [...prev, { id: `${Date.now()}-cancel`, type: "status", message: "Run cancelled" }]);
    setBusy(false);
  };
  const clear  = () => { if (busy) cancel(); setMessages([]); setAgentEvents([]); setError(null); };

  const handleKey = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); send(input); }
  };

  const charCount = input.length;
  const charWarn  = charCount > 420;
  const charOver  = charCount > 500;

  return (
    <>
      {/* ── FAB ─────────────────────────────────────────────────────────── */}
      <button
        onClick={() => { setOpen((o) => !o); setMinimized(false); }}
        aria-label={open ? "Close AI assistant" : "Open AI assistant"}
        className={cn(
          "fixed bottom-20 right-5 z-50 flex size-14 items-center justify-center rounded-full shadow-xl transition-all duration-200 md:bottom-7",
          "bg-primary text-primary-foreground hover:scale-105 hover:shadow-2xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
          open && "scale-95 shadow-lg",
        )}
      >
        <span className={cn("absolute inset-0 rounded-full bg-primary/30 animate-ping", !open && "opacity-0")} />
        <span className={cn("transition-transform duration-200", open && "rotate-90")}>
          {open ? <IconX /> : <IconChat />}
        </span>
        {!open && messages.length > 0 && (
          <span className="absolute -right-0.5 -top-0.5 flex size-4 items-center justify-center rounded-full bg-nust-gold text-[9px] font-bold text-black">
            {messages.filter((m) => m.role === "assistant").length}
          </span>
        )}
      </button>

      {/* ── Panel ───────────────────────────────────────────────────────── */}
      {open && (
        <div className={cn(
          "fixed bottom-[88px] right-5 z-50 flex flex-col overflow-hidden",
          "w-[min(96vw,480px)] rounded-2xl border border-border/60",
          "bg-background shadow-[0_24px_64px_-8px_rgba(0,0,0,0.22),0_0_0_1px_rgba(0,0,0,0.04)]",
          "transition-all duration-200 md:bottom-24",
          minimized ? "h-auto" : "h-[min(80vh,680px)]",
        )}>

          {/* ── Header ──────────────────────────────────────────────────── */}
          <div className="relative flex shrink-0 items-center justify-between overflow-hidden rounded-t-2xl bg-primary px-4 py-3.5">
            <div className="pointer-events-none absolute inset-0 opacity-[0.07]"
              style={{ backgroundImage: "radial-gradient(circle, currentColor 1px, transparent 1px)", backgroundSize: "18px 18px", color: "white" }}
            />
            <div className="relative flex items-center gap-3">
              <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-white/20 ring-2 ring-white/30">
                <svg className="size-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.8}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                </svg>
              </div>
              <div>
                <p className="text-sm font-semibold text-white">NUST Devs AI</p>
                <p className="flex items-center gap-1 text-[11px] text-white/70">
                  <span className={cn("size-1.5 rounded-full", busy ? "bg-amber-300 animate-pulse" : "bg-emerald-400")} />
                  {busy ? latestAgentMessage(agentEvents) : "Ready to help"}
                </p>
              </div>
            </div>
            <div className="relative flex items-center gap-0.5">
              {messages.length > 0 && (
                <button onClick={clear} title="Clear conversation"
                  className="rounded-lg p-2 text-white/60 transition-colors hover:bg-white/10 hover:text-white">
                  <IconTrash />
                </button>
              )}
              <button onClick={() => setMinimized((m) => !m)} title={minimized ? "Expand" : "Minimise"}
                className="rounded-lg p-2 text-white/60 transition-colors hover:bg-white/10 hover:text-white">
                <IconMinus />
              </button>
              <button onClick={() => setOpen(false)} title="Close"
                className="rounded-lg p-2 text-white/60 transition-colors hover:bg-white/10 hover:text-white">
                <IconX />
              </button>
            </div>
          </div>

          {!minimized && (
            <>
              {/* ── Messages ──────────────────────────────────────────── */}
              <div className="flex-1 overflow-y-auto overscroll-contain px-4 py-4 min-h-0 space-y-5 scroll-smooth">
                {agentEvents.length > 0 && (
                  <div className="rounded-xl border border-border/70 bg-muted/30 p-3">
                    <div className="mb-2 flex items-center justify-between">
                      <p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-muted-foreground">
                        Agent Activity
                      </p>
                      {busy && <TypingDots />}
                    </div>
                    <div className="space-y-1.5">
                      {agentEvents.slice(-6).map((event) => (
                        <div key={event.id} className="flex items-start justify-between gap-3 text-xs">
                          <div className="min-w-0">
                            <p className="font-medium text-foreground">{event.message}</p>
                            {event.toolName && (
                              <p className="truncate font-mono text-[11px] text-muted-foreground">{event.toolName}</p>
                            )}
                          </div>
                          <div className="shrink-0 text-[11px] text-muted-foreground">
                            {event.latencyMS ? `${event.latencyMS}ms` : event.success === false ? "failed" : ""}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {messages.length === 0 ? (
                  <div className="flex h-full flex-col items-center justify-center gap-6 py-4">
                    <div className="flex size-16 items-center justify-center rounded-2xl bg-primary/10">
                      <svg className="size-8 text-primary" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                      </svg>
                    </div>
                    <div className="space-y-1 text-center">
                      <p className="text-sm font-semibold">Ask me anything</p>
                      <p className="max-w-[280px] text-xs text-muted-foreground leading-relaxed">
                        Find developers, compare stats, discover trending projects, or get recruiter-ready insights.
                      </p>
                    </div>
                    <div className="grid w-full grid-cols-2 gap-2">
                      {SUGGESTED.map((s) => (
                        <button key={s.prompt} onClick={() => send(s.prompt)}
                          className="group flex flex-col gap-1 rounded-xl border border-border/80 bg-card/80 px-3 py-2.5 text-left transition-all hover:border-primary/40 hover:bg-primary/5 hover:shadow-sm">
                          <span className="text-base leading-none">{s.icon}</span>
                          <span className="text-[11px] font-medium text-foreground/80 group-hover:text-foreground">{s.label}</span>
                        </button>
                      ))}
                    </div>
                  </div>
                ) : (
                  messages.map((msg, i) => (
                    <div key={i} className={cn("group flex gap-2.5", msg.role === "user" ? "flex-row-reverse" : "flex-row")}>
                      {msg.role === "assistant" && (
                        <div className="mt-1 flex size-7 shrink-0 items-center justify-center rounded-full bg-primary/10 ring-1 ring-primary/20">
                          <svg className="size-3.5 text-primary" viewBox="0 0 24 24" fill="currentColor">
                            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 14.5v-9l6 4.5-6 4.5z" />
                          </svg>
                        </div>
                      )}

                      <div className={cn("flex min-w-0 flex-1 flex-col gap-1", msg.role === "user" && "items-end")}>
                        <div className={cn(
                          "rounded-2xl px-3.5 py-2.5 text-sm",
                          msg.role === "user"
                            ? "max-w-[82%] rounded-tr-sm bg-primary text-primary-foreground"
                            : "w-full rounded-tl-sm bg-muted/50 text-foreground ring-1 ring-border/50",
                        )}>
                          {msg.content
                            ? (msg.role === "assistant"
                                ? <RichMessage content={msg.content} />
                                : <p className="leading-relaxed">{msg.content}</p>)
                            : msg.streaming ? <TypingDots /> : null}
                          {msg.streaming && msg.content && (
                            <span className="ml-0.5 inline-block size-1.5 translate-y-[-1px] rounded-full bg-current animate-pulse" />
                          )}
                        </div>

                        <div className="flex items-center gap-1">
                          <span className="text-[10px] text-muted-foreground/60">{formatTime(msg.ts)}</span>
                          {msg.role === "assistant" && msg.content && !msg.streaming && (
                            <CopyButton text={msg.content} />
                          )}
                        </div>
                      </div>
                    </div>
                  ))
                )}

                {error && (
                  <div className="flex items-center gap-2 rounded-xl border border-destructive/20 bg-destructive/8 px-3 py-2.5 text-xs text-destructive">
                    <svg className="size-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
                      <circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/>
                    </svg>
                    {error}
                  </div>
                )}
                <div ref={bottomRef} />
              </div>

              {/* ── Input ─────────────────────────────────────────────── */}
              <div className="shrink-0 border-t border-border/60 bg-background/95 px-4 pb-4 pt-3">
                <div className={cn(
                  "flex items-end gap-2 rounded-xl border bg-background px-3 py-2 transition-colors",
                  "focus-within:border-primary/60 focus-within:ring-1 focus-within:ring-primary/30",
                  charOver ? "border-destructive" : "border-border",
                )}>
                  <textarea ref={inputRef} rows={1} value={input}
                    onChange={(e) => setInput(e.target.value)} onKeyDown={handleKey}
                    placeholder="Ask about developers, stats, projects… (Enter to send)"
                    maxLength={520} disabled={busy}
                    className="flex-1 resize-none bg-transparent text-sm outline-none placeholder:text-muted-foreground/60 disabled:opacity-50 min-h-[22px] max-h-[120px] py-0.5 leading-relaxed"
                  />
                  {busy ? (
                    <button type="button" onClick={cancel} title="Stop"
                      className="mb-0.5 flex shrink-0 size-8 items-center justify-center rounded-lg bg-muted text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive">
                      <IconStop />
                    </button>
                  ) : (
                    <button type="button" onClick={() => send(input)} disabled={!input.trim() || charOver} title="Send (Enter)"
                      className="mb-0.5 flex shrink-0 size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground transition-all hover:opacity-90 disabled:opacity-30">
                      <IconSend />
                    </button>
                  )}
                </div>
                <div className="mt-2 flex items-center justify-between">
                  <p className="text-[10px] text-muted-foreground/50">
                    Live NUST developer data · Shift+Enter for newline
                  </p>
                  {charWarn && (
                    <span className={cn("text-[10px] tabular-nums", charOver ? "text-destructive" : "text-muted-foreground")}>
                      {charCount}/500
                    </span>
                  )}
                </div>
              </div>
            </>
          )}
        </div>
      )}
    </>
  );
}

function describeAgentEvent(event: ChatAgentEvent) {
  if (event.type === "tool_call") return `Running ${event.tool_name ?? "tool"}`;
  if (event.type === "tool_done") {
    if (event.success === false) return `${event.tool_name ?? "Tool"} failed`;
    return `${event.tool_name ?? "Tool"} finished`;
  }
  return event.message ?? "Thinking";
}

function latestAgentMessage(events: AgentTimelineEvent[]) {
  if (events.length === 0) return "Thinking…";
  return events[events.length - 1]?.message ?? "Thinking…";
}
