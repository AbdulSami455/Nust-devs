import type { Metadata } from "next";
import { ProfileClient } from "./client";

const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function generateMetadata(
  { params }: { params: Promise<{ username: string }> }
): Promise<Metadata> {
  const { username } = await params;
  try {
    const res = await fetch(`${BASE}/api/v1/developers/${username}`, { next: { revalidate: 300 } });
    if (!res.ok) throw new Error();
    const dev = await res.json();
    const name = dev.display_name ?? dev.github_username;
    const description = dev.bio ?? `${name}'s GitHub stats on NUST Devs`;
    return {
      title: `${name} (@${dev.github_username}) — NUST Devs`,
      description,
      openGraph: {
        title: `${name} — NUST Devs`,
        description,
        images: dev.avatar_url ? [{ url: dev.avatar_url, width: 400, height: 400 }] : [],
        type: "profile",
      },
      twitter: {
        card: "summary",
        title: `${name} — NUST Devs`,
        description,
        images: dev.avatar_url ? [dev.avatar_url] : [],
      },
    };
  } catch {
    return {
      title: `@${username} — NUST Devs`,
      description: `${username}'s developer profile on NUST Devs`,
    };
  }
}

export default async function DeveloperProfilePage(
  { params }: { params: Promise<{ username: string }> }
) {
  const { username } = await params;
  return <ProfileClient username={username} />;
}
