"use client";

import { useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { ArrowLeft, CheckCircle2, GitPullRequest, GraduationCap } from "lucide-react";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";

const courseOptions = ["SE", "CS", "AI", "DS", "EE", "Other"];

export function JoinClient() {
  const [form, setForm] = useState({
    github_username: "",
    email: "",
    display_name: "",
    batch: "",
    course: "",
    other_course: "",
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
    const course =
      form.course === "Other" ? form.other_course.trim() || "Other" : form.course.trim();
    setSubmitting(true);
    try {
      await api.public.submitProfileRequest({
        github_username: form.github_username.trim(),
        ...(form.email.trim() ? { email: form.email.trim() } : {}),
        ...(form.display_name.trim() ? { display_name: form.display_name.trim() } : {}),
        ...(form.batch.trim() ? { batch: form.batch.trim() } : {}),
        ...(course ? { course } : {}),
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
        <div className="mx-auto mb-5 flex size-12 items-center justify-center rounded-full bg-primary/10 text-primary">
          <CheckCircle2 className="size-6" />
        </div>
        <h1 className="text-2xl font-semibold tracking-tight">Request received</h1>
        <p className="mt-3 leading-6 text-muted-foreground">
          An admin will review your profile request. Once approved, your GitHub stats will appear on
          NUST Devs.
        </p>
        <Link href="/" className="mt-6 inline-flex items-center gap-2 text-sm font-medium text-primary hover:underline">
          <ArrowLeft className="size-4" />
          Back to home
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto grid max-w-6xl gap-8 px-4 py-10 sm:px-6 sm:py-14 lg:grid-cols-[0.85fr_1.15fr] lg:items-start">
      <aside className="lg:sticky lg:top-24">
        <Badge variant="secondary" className="mb-5 gap-2 rounded-full px-3 py-1">
          <GraduationCap className="size-3.5" />
          NUST developer community
        </Badge>
        <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">Join NUST Devs</h1>
        <p className="mt-4 max-w-md text-sm leading-6 text-muted-foreground">
          Submit your GitHub profile for review. Academic details are optional, but they help admins
          recognize batches and departments more quickly.
        </p>
        <div className="mt-8 hidden rounded-lg border bg-card p-5 lg:block">
          <div className="flex items-center gap-3">
            <Image
              src="/nust-logo.svg"
              alt="NUST"
              width={44}
              height={44}
              className="size-11 rounded-full ring-1 ring-border"
            />
            <div>
              <p className="font-medium">Reviewed before publishing</p>
              <p className="text-sm text-muted-foreground">Each profile appears after admin approval.</p>
            </div>
          </div>
        </div>
      </aside>

      <form onSubmit={handleSubmit} className="bento-card space-y-6">
        <section className="space-y-4">
          <div className="flex items-center gap-2">
            <GitPullRequest className="size-4 text-primary" />
            <h2 className="text-base font-semibold">GitHub identity</h2>
          </div>
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

          <div className="grid gap-4 sm:grid-cols-2">
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
                placeholder="Name shown to admins"
                value={form.display_name}
                onChange={(e) => setForm((f) => ({ ...f, display_name: e.target.value }))}
              />
            </div>
          </div>
        </section>

        <section className="space-y-4 border-t pt-6">
          <div className="flex items-center gap-2">
            <GraduationCap className="size-4 text-primary" />
            <h2 className="text-base font-semibold">Academic details</h2>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="batch">Batch (optional)</Label>
              <Input
                id="batch"
                placeholder="e.g. 2024"
                value={form.batch}
                onChange={(e) => setForm((f) => ({ ...f, batch: e.target.value }))}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="course">Course (optional)</Label>
              <select
                id="course"
                value={form.course}
                onChange={(e) =>
                  setForm((f) => ({ ...f, course: e.target.value, other_course: "" }))
                }
                className="h-8 w-full rounded-lg border border-input bg-background px-2.5 py-1 text-sm outline-none transition-colors focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
              >
                <option value="">Select course</option>
                {courseOptions.map((course) => (
                  <option key={course} value={course}>
                    {course}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {form.course === "Other" && (
            <div className="space-y-2">
              <Label htmlFor="other-course">Other course (optional)</Label>
              <Input
                id="other-course"
                placeholder="Enter course"
                value={form.other_course}
                onChange={(e) => setForm((f) => ({ ...f, other_course: e.target.value }))}
              />
            </div>
          )}
        </section>

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
