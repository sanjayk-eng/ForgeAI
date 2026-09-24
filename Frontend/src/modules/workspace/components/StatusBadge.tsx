import type { WorkspaceInvite } from "../types/workspace.types";

const styles: Record<string, string> = {
  PENDING: "border-forge-signal/25 bg-forge-signal/[0.08] text-forge-signal",
  ACCEPTED: "border-forge-accent/25 bg-forge-accent/[0.08] text-forge-accent",
  REJECTED: "border-[var(--destructive)]/25 bg-[var(--destructive)]/[0.08] text-[var(--destructive)]",
  EXPIRED: "border-white/10 bg-white/[0.04] text-forge-muted",
  REVOKED: "border-white/10 bg-white/[0.04] text-forge-muted",
};

export function StatusBadge({ status }: { status: WorkspaceInvite["status"] | string }) {
  const normalized = status.toUpperCase();
  return (
    <span className={`w-fit rounded-full border px-2.5 py-1 font-mono text-[10px] font-bold uppercase tracking-wide ${styles[normalized] ?? styles.PENDING}`}>
      {normalized}
    </span>
  );
}
