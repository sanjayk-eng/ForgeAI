import { AlertCircle, RefreshCw, Users } from "lucide-react";
import type { WorkspaceMember } from "../api";

type Props = {
  members: WorkspaceMember[];
  loading: boolean;
  error: boolean;
  onRetry: () => void;
  onRoleChange: (userId: string, role: string) => void;
  onRemove: (userId: string) => void;
  roleUpdating: boolean;
  removing: boolean;
};

const roleOptions = ["OWNER", "ADMIN", "MEMBER"];

export function WorkspaceMemberTable({
  members,
  loading,
  error,
  onRetry,
  onRoleChange,
  onRemove,
  roleUpdating,
  removing,
}: Props) {
  if (loading) return <TableMessage label="Loading workspace members..." />;
  if (error) {
    return (
      <div className="grid min-h-[190px] place-content-center justify-items-center gap-3 text-center">
        <AlertCircle size={22} className="text-forge-signal" />
        <p className="m-0 text-sm text-forge-soft">
          Members could not be loaded.
        </p>
        <button
          className="inline-flex items-center gap-2 border border-white/[0.12] px-3 py-2 text-xs font-bold text-forge-soft transition hover:border-forge-accent/40 hover:text-forge-text"
          onClick={onRetry}
        >
          <RefreshCw size={13} /> Try again
        </button>
      </div>
    );
  }
  if (members.length === 0)
    return (
      <TableMessage icon label="No members have joined this workspace yet." />
    );

  return (
    <div>
      <div className="hidden min-w-0 grid-cols-[minmax(0,1fr)_auto_auto_auto] gap-2 border-b border-white/[0.08] px-4 py-3 font-mono text-[10px] uppercase tracking-[.1em] text-forge-muted sm:grid sm:px-6">
        <span>Member</span>
        <span>Role</span>
        <span>Joined</span>
        <span />
      </div>
      {members.map((member) => (
        <MemberRow
          key={member.id}
          member={member}
          onRoleChange={onRoleChange}
          onRemove={onRemove}
          roleUpdating={roleUpdating}
          removing={removing}
        />
      ))}
    </div>
  );
}

function MemberRow({
  member,
  onRoleChange,
  onRemove,
  roleUpdating,
  removing,
}: {
  member: WorkspaceMember;
  onRoleChange: Props["onRoleChange"];
  onRemove: Props["onRemove"];
  roleUpdating: boolean;
  removing: boolean;
}) {
  const isOwner = member.role === "OWNER";
  return (
    <div className="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-x-3 gap-y-3 border-b border-white/[0.06] p-4 last:border-0 sm:grid-cols-[minmax(0,1fr)_auto_auto_auto] sm:gap-2 sm:px-6 sm:py-4">
      <div className="flex min-w-0 items-center gap-3">
        <span className="grid size-9 shrink-0 place-items-center bg-[#8ed4ff] font-mono text-xs font-bold text-forge-bg">
          {(member.user.name || member.user.email || member.user_id)
            .slice(0, 2)
            .toUpperCase()}
        </span>
        <div className="min-w-0">
          <strong className="block truncate text-[13px] text-forge-text">
            {member.user.name || member.user.email}
          </strong>
          <small className="mt-1 block truncate font-mono text-[10px] text-forge-muted">
            {member.user.email} · {member.user_id.slice(0, 8)}
          </small>
        </div>
      </div>
      <select
        className="w-[92px] rounded border border-white/[0.11] bg-forge-card p-2 font-mono text-[11px] text-forge-soft outline-none focus:border-forge-accent disabled:cursor-not-allowed disabled:opacity-60 sm:w-[100px]"
        value={member.role}
        disabled={isOwner || roleUpdating}
        onChange={(event) => onRoleChange(member.user_id, event.target.value)}
        aria-label={`Role for ${member.user_id}`}
      >
        {roleOptions.map((role) => (
          <option key={role}>{role}</option>
        ))}
      </select>
      <span className="hidden whitespace-nowrap font-mono text-[10px] text-forge-muted sm:block">
        {new Date(member.created_at).toLocaleDateString()}
      </span>
      <button
        className="col-start-2 row-start-1 whitespace-nowrap text-[11px] font-bold text-forge-muted transition hover:text-[#ff8e7a] disabled:cursor-not-allowed disabled:opacity-30 sm:col-auto sm:row-auto"
        disabled={isOwner || removing}
        onClick={() => onRemove(member.user_id)}
      >
        {removing ? "Removing..." : isOwner ? "Owner" : "Remove"}
      </button>
    </div>
  );
}

function TableMessage({
  label,
  icon = false,
}: {
  label: string;
  icon?: boolean;
}) {
  return (
    <div className="grid min-h-[190px] place-content-center justify-items-center gap-3 text-center text-sm text-forge-muted">
      {icon && <Users size={24} />}
      {label}
    </div>
  );
}
