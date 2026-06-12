"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { api, type Developer, type DeveloperRequest } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

interface SyncStatus {
  github_rate_limit: { remaining: number; reset_at: string };
}

export default function DashboardPage() {
  const router = useRouter();
  const [developers, setDevelopers] = useState<Developer[]>([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [form, setForm] = useState({ github_username: "", email: "", display_name: "", notes: "" });
  const [submitting, setSubmitting] = useState(false);
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [syncing, setSyncing] = useState(false);
  const [pendingRequests, setPendingRequests] = useState<DeveloperRequest[]>([]);

  const fetchDevelopers = useCallback(async () => {
    try {
      const data = await api.developers.list();
      setDevelopers(data);
    } catch (err: unknown) {
      if (err instanceof Error && err.message === "HTTP 401") {
        router.push("/admin");
      } else {
        toast.error("Failed to load developers");
      }
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    let cancelled = false;

    api.developers
      .list()
      .then((data) => {
        if (!cancelled) setDevelopers(data);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        if (err instanceof Error && err.message === "HTTP 401") {
          router.push("/admin");
        } else {
          toast.error("Failed to load developers");
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    api.admin.sync
      .status()
      .then((s) => {
        if (!cancelled) setSyncStatus(s as SyncStatus);
      })
      .catch(() => {});
    api.admin.profileRequests
      .list("pending")
      .then((requests) => {
        if (!cancelled) setPendingRequests(requests);
      })
      .catch(() => {});

    return () => {
      cancelled = true;
    };
  }, [router]);

  async function handleApproveRequest(id: string, username: string) {
    try {
      await api.admin.profileRequests.approve(id);
      toast.success(`Approved @${username}`);
      setPendingRequests((prev) => prev.filter((r) => r.id !== id));
      fetchDevelopers();
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : "Failed to approve");
    }
  }

  async function handleRejectRequest(id: string, username: string) {
    if (!confirm(`Reject request for @${username}?`)) return;
    try {
      await api.admin.profileRequests.reject(id);
      toast.success(`Rejected @${username}`);
      setPendingRequests((prev) => prev.filter((r) => r.id !== id));
    } catch {
      toast.error("Failed to reject request");
    }
  }

  async function handleSyncAll() {
    setSyncing(true);
    try {
      await api.admin.sync.trigger();
      toast.success("Sync triggered for all developers");
      api.admin.sync.status().then((s) => setSyncStatus(s as SyncStatus)).catch(() => {});
    } catch {
      toast.error("Failed to trigger sync");
    } finally {
      setSyncing(false);
    }
  }

  async function handleAdd(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSubmitting(true);
    try {
      await api.developers.create({
        github_username: form.github_username.trim(),
        ...(form.email.trim() ? { email: form.email.trim() } : {}),
        ...(form.display_name.trim() ? { display_name: form.display_name.trim() } : {}),
        ...(form.notes.trim() ? { notes: form.notes.trim() } : {}),
      });
      toast.success("Developer added");
      setOpen(false);
      setForm({ github_username: "", email: "", display_name: "", notes: "" });
      fetchDevelopers();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to add developer";
      toast.error(msg.includes("409") || msg.toLowerCase().includes("already") ? "This GitHub profile is already registered" : msg);
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(id: string, username: string) {
    if (!confirm(`Delete ${username}?`)) return;
    try {
      await api.developers.delete(id);
      toast.success("Developer removed");
      setDevelopers((prev) => prev.filter((d) => d.id !== id));
    } catch {
      toast.error("Failed to delete developer");
    }
  }

  function handleLogout() {
    localStorage.removeItem("token");
    router.push("/admin");
  }

  function field(key: keyof typeof form) {
    return (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [key]: e.target.value }));
  }

  return (
    <div className="min-h-screen bg-muted/40">
      <header className="border-b bg-background px-6 py-4 flex items-center justify-between">
        <h1 className="text-xl font-bold">NUST Devs — Admin</h1>
        <Button variant="outline" size="sm" onClick={handleLogout}>
          Logout
        </Button>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-8 space-y-6">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Total Developers</CardTitle>
            </CardHeader>
            <CardContent><p className="text-3xl font-bold">{developers.length}</p></CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Synced</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold">{developers.filter((d) => d.last_synced_at).length}</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Total Stars</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold">{developers.reduce((s, d) => s + d.total_stars, 0)}</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">GitHub Quota</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold">
                {syncStatus ? syncStatus.github_rate_limit.remaining : "—"}
              </p>
              {syncStatus && (
                <p className="text-xs text-muted-foreground mt-1">
                  resets {new Date(syncStatus.github_rate_limit.reset_at).toLocaleTimeString()}
                </p>
              )}
            </CardContent>
          </Card>
        </div>

        {pendingRequests.length > 0 && (
          <section className="space-y-3">
            <h2 className="text-lg font-semibold">
              Pending profile requests{" "}
              <Badge variant="secondary">{pendingRequests.length}</Badge>
            </h2>
            <div className="rounded-lg border bg-background overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>GitHub</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead>Batch</TableHead>
                    <TableHead>Course</TableHead>
                    <TableHead>Message</TableHead>
                    <TableHead>Submitted</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {pendingRequests.map((req) => (
                    <TableRow key={req.id}>
                      <TableCell className="font-mono font-medium">@{req.github_username}</TableCell>
                      <TableCell className="text-muted-foreground">{req.email || "—"}</TableCell>
                      <TableCell>{req.display_name || "—"}</TableCell>
                      <TableCell>{req.batch || "—"}</TableCell>
                      <TableCell>{req.course || "—"}</TableCell>
                      <TableCell className="max-w-[200px] truncate text-sm text-muted-foreground">
                        {req.message || "—"}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {new Date(req.created_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell className="space-x-2 text-right">
                        <Button size="sm" onClick={() => handleApproveRequest(req.id, req.github_username)}>
                          Approve
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          className="text-destructive"
                          onClick={() => handleRejectRequest(req.id, req.github_username)}
                        >
                          Reject
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </section>
        )}

        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Developers</h2>
          <Button variant="outline" size="sm" onClick={handleSyncAll} disabled={syncing}>
            {syncing ? "Syncing…" : "Sync All"}
          </Button>
          <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger
              className="inline-flex items-center justify-center gap-2 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            >
              + Add Developer
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Register Developer</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleAdd} className="space-y-4 pt-2">
                <div className="space-y-1">
                  <Label htmlFor="username">GitHub Username *</Label>
                  <Input id="username" required value={form.github_username} onChange={field("github_username")} />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="dev-email">Email</Label>
                  <Input id="dev-email" type="email" value={form.email} onChange={field("email")} />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="display-name">Display Name</Label>
                  <Input id="display-name" value={form.display_name} onChange={field("display_name")} />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="notes">Notes</Label>
                  <Input id="notes" value={form.notes} onChange={field("notes")} />
                </div>
                <Button type="submit" className="w-full" disabled={submitting}>
                  {submitting ? "Adding…" : "Add Developer"}
                </Button>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        <div className="rounded-lg border bg-background overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Username</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Display Name</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Stars</TableHead>
                <TableHead>Repos</TableHead>
                <TableHead>Last Synced</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center text-muted-foreground py-8">Loading…</TableCell>
                </TableRow>
              ) : developers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center text-muted-foreground py-8">
                    No developers yet. Add one to get started.
                  </TableCell>
                </TableRow>
              ) : (
                developers.map((dev) => (
                  <TableRow key={dev.id}>
                    <TableCell className="font-mono font-medium">
                      <a href={`https://github.com/${dev.github_username}`} target="_blank" rel="noopener noreferrer" className="hover:underline">
                        {dev.github_username}
                      </a>
                    </TableCell>
                    <TableCell className="text-muted-foreground">{dev.email || "—"}</TableCell>
                    <TableCell>{dev.display_name ?? "—"}</TableCell>
                    <TableCell>
                      <Badge variant={dev.verification_status === "registered" ? "secondary" : "default"}>
                        {dev.verification_status}
                      </Badge>
                    </TableCell>
                    <TableCell>{dev.total_stars}</TableCell>
                    <TableCell>{dev.public_repos}</TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {dev.last_synced_at ? new Date(dev.last_synced_at).toLocaleDateString() : "Never"}
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-destructive hover:text-destructive"
                        onClick={() => handleDelete(dev.id, dev.github_username)}
                      >
                        Remove
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </main>
    </div>
  );
}
