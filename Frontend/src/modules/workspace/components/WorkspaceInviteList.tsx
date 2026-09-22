import { Clock3, Mail, Send, X } from "lucide-react";
import type { WorkspaceInvite } from "../api";

type Props = {
  invites: WorkspaceInvite[];
  loading: boolean;
  error: boolean;
  onRetry: () => void;
  onRevoke?: (inviteId: string) => void;
  revoking?: boolean;
};

const statusStyles: Record<string, string> = {
  PENDING: "border-forge-signal/25 bg-forge-signal/[0.08] text-forge-signal",
  ACCEPTED: "border-forge-accent/25 bg-forge-accent/[0.08] text-[#bde986]",
  REJECTED: "border-[#ff8e7a]/25 bg-[#ff8e7a]/[0.08] text-[#ff9c8c]",
  EXPIRED: "border-white/10 bg-white/[0.04] text-forge-muted",
  REVOKED: "border-white/10 bg-white/[0.04] text-forge-muted",
};

export function WorkspaceInviteList({
  invites,
  loading,
  error,
  onRetry,
  onRevoke,
  revoking = false,
}: Props) {
  if (loading) return <InviteMessage label="Loading invitations..." />;
  if (error)
    return (
      <InviteMessage
        label="Invitations could not be loaded."
        action={onRetry}
      />
    );
  if (invites.length === 0)
    return <InviteMessage label="No invitations sent yet." />;

  return (
    <div className="divide-y divide-white/[0.06]">
      {invites.map((invite) => {
        const status = invite.status.toUpperCase();
        const canRevoke = status === "PENDING" && onRevoke;
        return (
          <div
            key={invite.id}
            className="grid gap-3 px-5 py-4 sm:grid-cols-[minmax(0,1fr)_110px_150px_auto] sm:items-center sm:px-6"
          >
            <div className="flex min-w-0 items-center gap-3">
              <span className="grid size-8 shrink-0 place-items-center border border-white/[0.1] text-forge-muted">
                <Mail size={15} />
              </span>
              <div className="min-w-0">
                <p className="m-0 truncate text-sm font-bold text-forge-text">
                  {invite.email}
                </p>
                <p className="mt-1 font-mono text-[10px] uppercase text-forge-muted">
                  Invited as {invite.role} · by{" "}
                  {invite.invited_by_user.name || invite.invited_by_user.email}
                </p>
              </div>
            </div>
            <span
              className={`w-fit border px-2 py-1 font-mono text-[10px] uppercase ${statusStyles[status] ?? statusStyles.PENDING}`}
            >
              {status}
            </span>
            <span className="flex items-center gap-2 font-mono text-[10px] text-forge-muted">
              <Clock3 size={13} />
              {status === "PENDING"
                ? `Expires ${new Date(invite.expires_at).toLocaleDateString()}`
                : `Sent ${new Date(invite.created_at).toLocaleDateString()}`}
            </span>
            {canRevoke && (
              <button
                onClick={() => onRevoke(invite.id)}
                disabled={revoking}
                className="flex items-center gap-1.5 text-xs font-medium text-forge-muted transition hover:text-red-400 disabled:opacity-50"
                title="Revoke invitation"
              >
                <X size={14} />
                Revoke
              </button>
            )}
          </div>
        );
      })}
    </div>
  );
}

function InviteMessage({
  label,
  action,
}: {
  label: string;
  action?: () => void;
}) {
  return (
    <div className="grid min-h-[130px] place-content-center justify-items-center gap-3 text-center text-sm text-forge-muted">
      <Send size={20} />
      {label}
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
