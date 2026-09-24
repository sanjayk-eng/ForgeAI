import { AlertTriangle, Clock3, Mail, Send, X } from "lucide-react";
import { useEffect, useState } from "react";
import type { WorkspaceInvite } from "../types/workspace.types";
import { StatusBadge } from "./StatusBadge";

type Props = {
  invites: WorkspaceInvite[];
  loading: boolean;
  error: boolean;
  onRetry: () => void;
  onRevoke?: (inviteId: string) => void;
  revoking?: boolean;
};

export function WorkspaceInviteList({
  invites,
  loading,
  error,
  onRetry,
  onRevoke,
  revoking = false,
}: Props) {
  const [pendingRevoke, setPendingRevoke] = useState<WorkspaceInvite | null>(null);
  useEffect(() => {
    if (!pendingRevoke) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !revoking) setPendingRevoke(null);
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [pendingRevoke, revoking]);
  if (loading) return <InviteSkeleton />;
  if (error)
    return (
      <InviteMessage
        label="Invitations could not be loaded."
        action={onRetry}
      />
    );
  if (invites.length === 0)
    return <InviteMessage label="No pending invitations" description="Invitations you send will appear here." />;

  return (
    <div className="divide-y divide-[var(--border)]">
      {invites.map((invite) => {
        const status = invite.status.toUpperCase();
        const canRevoke = status === "PENDING" && onRevoke;
        return (
          <div
            key={invite.id}
            className="group grid gap-4 px-5 py-5 transition hover:bg-white/[0.025] sm:grid-cols-[minmax(0,1fr)_auto_170px_auto] sm:items-center sm:px-6"
          >
            <div className="flex min-w-0 items-center gap-3">
              <span className="grid size-9 shrink-0 place-items-center rounded-lg border border-[var(--border)] bg-white/[0.03] text-forge-muted">
                <Mail size={15} />
              </span>
              <div className="min-w-0">
                <p className="m-0 truncate text-sm font-bold text-forge-text">
                  {invite.email}
                </p>
                <p className="mt-1 text-[11px] text-forge-muted">
                  Invited as {invite.role} · by{" "}
                  {invite.invited_by_user?.name || invite.invited_by_user?.email || "Unknown"}
                </p>
              </div>
            </div>
            <span
              className="w-fit"
            >
              <StatusBadge status={status} />
            </span>
            <span className="flex items-center gap-2 text-[11px] text-forge-muted">
              <Clock3 size={13} />
              {status === "PENDING"
                ? `Expires ${new Date(invite.expires_at).toLocaleDateString()}`
                : `Sent ${new Date(invite.created_at).toLocaleDateString()}`}
            </span>
            {canRevoke && (
              <button
                onClick={() => setPendingRevoke(invite)}
                disabled={revoking}
                className="flex items-center gap-1.5 text-xs font-semibold text-forge-muted opacity-70 transition hover:text-[var(--destructive)] hover:opacity-100 disabled:opacity-50"
                title="Revoke invitation"
              >
                <X size={14} />
                Revoke
              </button>
            )}
          </div>
        );
      })}
      {pendingRevoke && onRevoke && (
        <div className="fixed inset-0 z-[60] grid place-items-center bg-black/70 p-5 backdrop-blur-sm" role="dialog" aria-modal="true" aria-labelledby="revoke-title" onClick={() => !revoking && setPendingRevoke(null)}>
          <div className="w-full max-w-[420px] rounded-xl border border-[var(--border)] bg-forge-card p-6 shadow-2xl" onClick={(event) => event.stopPropagation()}>
            <div className="flex size-10 items-center justify-center rounded-lg border border-[var(--destructive)]/25 bg-[var(--destructive)]/10 text-[var(--destructive)]"><AlertTriangle size={19} /></div>
            <h2 id="revoke-title" className="mt-5 text-lg font-bold text-forge-text">Revoke invitation?</h2>
            <p className="mt-2 text-sm leading-6 text-forge-muted">The invitation for <strong className="font-semibold text-forge-soft">{pendingRevoke.email}</strong> will no longer be usable.</p>
            <div className="mt-7 flex justify-end gap-2">
              <button className="rounded-lg border border-[var(--border)] px-4 py-2.5 text-xs font-bold text-forge-soft transition hover:bg-white/[0.05] focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent" onClick={() => setPendingRevoke(null)} disabled={revoking}>Cancel</button>
              <button className="rounded-lg bg-[var(--destructive)] px-4 py-2.5 text-xs font-bold text-[var(--destructive-foreground)] transition hover:brightness-110 disabled:cursor-wait disabled:opacity-50" onClick={() => { onRevoke(pendingRevoke.id); setPendingRevoke(null); }} disabled={revoking}>{revoking ? "Revoking..." : "Revoke invitation"}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function InviteMessage({
  label,
  action,
  description,
}: {
  label: string;
  action?: () => void;
  description?: string;
}) {
  return (
    <div className="grid min-h-[130px] place-content-center justify-items-center gap-3 text-center text-sm text-forge-muted">
      <Send size={20} />
      <strong className="text-forge-soft">{label}</strong>
      {description && <span className="text-xs text-forge-muted">{description}</span>}
      {action && (
        <button
          className="text-xs font-bold text-forge-accent"
          onClick={action}
        >
          Try again
        </button>
      )}
    </div>
  );
}

function InviteSkeleton() {
  return (
    <div className="divide-y divide-[var(--border)]" aria-label="Loading invitations">
      {[1, 2].map((item) => (
        <div key={item} className="flex items-center gap-3 px-5 py-5 sm:px-6">
          <span className="size-9 animate-pulse rounded-lg bg-white/[0.08]" />
          <span className="grid flex-1 gap-2"><span className="h-3 w-52 animate-pulse rounded bg-white/[0.08]" /><span className="h-2.5 w-40 animate-pulse rounded bg-white/[0.05]" /></span>
          <span className="hidden h-5 w-16 animate-pulse rounded-full bg-white/[0.08] sm:block" />
        </div>
      ))}
    </div>
  );
}
