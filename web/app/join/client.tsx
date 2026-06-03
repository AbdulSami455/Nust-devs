"use client";

import { useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";

export function JoinClient() {
  const [form, setForm] = useState({
    github_username: "",
    email: "",
    display_name: "",
    message: "",
  });
  const [checking, setChecking] = useState(false);
  const [availability, setAvailability] = useState<{
    available?: boolean;
    reason?: string;
    username?: string;
  } | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  async function checkUsername() {
    const u = form.github_username.trim();
    if (!u) {
      setAvailability(null);
      return;
    }
    setChecking(true);
    try {
      const res = await api.public.checkProfileUsername(u);
      setAvailability(res);
      if (res.username && res.username !== form.github_username) {
        setForm((f) => ({ ...f, github_username: res.username! }));
      }
    } catch {
      setAvailability(null);
    } finally {
      setChecking(false);
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    try {
      await api.public.submitProfileRequest({
        github_username: form.github_username.trim(),
        ...(form.email.trim() ? { email: form.email.trim() } : {}),
        ...(form.display_name.trim() ? { display_name: form.display_name.trim() } : {}),
        ...(form.message.trim() ? { message: form.message.trim() } : {}),
      });
      setSubmitted(true);
      toast.success("Request submitted for admin review");
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : "Could not submit request");
    } finally {
      setSubmitting(false);
    }
  }

  if (submitted) {
    return (
      <div className="mx-auto max-w-lg px-4 py-16 text-center sm:px-6">
        <h1 className="text-2xl font-bold">Request received</h1>
        <p className="mt-3 text-muted-foreground">
          An admin will review your profile request. Once approved, your GitHub stats will appear on
          NUST Devs.
        </p>
        <Link href="/" className="mt-6 inline-block text-primary hover:underline">
          Back to home
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-lg px-4 py-10 sm:px-6 sm:py-14">
      <Badge variant="secondary" className="mb-4">
        NUST developer community
      </Badge>
      <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">Join NUST Devs</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        Submit your GitHub username for admin approval. Each profile can only be listed once.
      </p>

      <form onSubmit={handleSubmit} className="mt-8 space-y-5">
        <div className="space-y-2">
          <Label htmlFor="username">GitHub username</Label>
          <div className="flex gap-2">
            <Input
              id="username"
              required
              placeholder="octocat"
              value={form.github_username}
              onChange={(e) => {
                setForm((f) => ({ ...f, github_username: e.target.value }));
                setAvailability(null);
              }}
              onBlur={checkUsername}
            />
            <Button type="button" variant="outline" onClick={checkUsername} disabled={checking}>
              {checking ? "…" : "Check"}
            </Button>
          </div>
          {availability && (
            <p
              className={`text-xs ${availability.available ? "text-emerald-600 dark:text-emerald-400" : "text-destructive"}`}
            >
              {availability.available
                ? `Available as @${availability.username ?? form.github_username}`
                : availability.reason === "already registered"
                  ? "This profile is already on NUST Devs."
                  : availability.reason === "request pending"
                    ? "A pending request already exists for this username."
                    : "Invalid GitHub username."}
            </p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="email">NUST email (optional)</Label>
          <Input
            id="email"
            type="email"
            placeholder="name@nust.edu.pk"
            value={form.email}
            onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="display">Display name (optional)</Label>
          <Input
            id="display"
            value={form.display_name}
            onChange={(e) => setForm((f) => ({ ...f, display_name: e.target.value }))}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="message">Note for admin (optional)</Label>
          <Input
            id="message"
            placeholder="e.g. SEECS student, open source contributor"
            value={form.message}
            onChange={(e) => setForm((f) => ({ ...f, message: e.target.value }))}
          />
        </div>

        <Button
          type="submit"
          className="w-full"
          disabled={submitting || availability?.available === false}
        >
          {submitting ? "Submitting…" : "Submit for approval"}
        </Button>
      </form>
    </div>
  );
}
