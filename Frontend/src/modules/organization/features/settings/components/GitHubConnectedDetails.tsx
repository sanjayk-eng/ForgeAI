import { useState } from "react";
import { CalendarDays, ExternalLink, GitBranch, LogOut, UserRound } from "lucide-react";
import type { OrganizationIntegration } from "../../../api/integrations.api";

export function GitHubConnectedDetails({
  integration,
  canManage,
  disconnecting,
  onDisconnect,
}: {
  integration: OrganizationIntegration;
  canManage: boolean;
  disconnecting: boolean;
  onDisconnect: () => void;
}) {
  const [confirming, setConfirming] = useState(false);
  const active = integration.status === "ACTIVE";
  const connectedDate = new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(integration.connected_at));

  return (
    <article className="border border-[var(--border)] bg-forge-panel/65 p-4 sm:p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <span className="grid size-10 shrink-0 place-items-center border border-[var(--border)] bg-[var(--surface-subtle)] text-forge-text">
            <GitBranch size={21} />
          </span>
          <div className="min-w-0">
            <h3 className="m-0 truncate text-sm font-bold text-forge-text">{integration.external_account_name}</h3>
            <p className="m-0 mt-1 text-xs text-forge-muted">GitHub {integration.external_account_type}</p>
          </div>
        </div>
        <span className={`inline-flex items-center gap-2 border px-2.5 py-1 text-[10px] font-bold ${active ? "border-emerald-500/20 bg-emerald-500/[0.08] text-emerald-500" : "border-[var(--border)] bg-[var(--surface-subtle)] text-forge-muted"}`}>
          <span className={`size-1.5 rounded-full ${active ? "bg-emerald-500" : "bg-current"}`} />
          {active ? "Connected" : integration.status.toLowerCase()}
        </span>
      </div>

      <dl className="mt-5 grid gap-x-6 gap-y-4 border-t border-[var(--border)] pt-4 sm:grid-cols-2">
        <Detail label="Connected account" value={integration.external_account_name} />
        <Detail label="Integration type" value="GitHub App" />
        <Detail label="Account type" value={integration.external_account_type} icon={<UserRound size={13} />} />
        <Detail label="Installation ID" value={`#${integration.installation_id}`} />
        <Detail label="Connected on" value={connectedDate} icon={<CalendarDays size={13} />} />
      </dl>

      {active && (
        <div className="mt-5 flex flex-wrap justify-end gap-2 border-t border-[var(--border)] pt-4">
          <p className="mr-auto text-xs text-forge-muted">Disconnecting only removes this ForgeAI organization connection; it does not uninstall the GitHub App.</p>
          <a
            className="inline-flex min-h-10 items-center gap-2 border border-[var(--border)] px-3 text-xs font-bold text-forge-text no-underline transition hover:bg-[var(--surface-hover)]"
            href={`https://github.com/settings/installations/${encodeURIComponent(integration.installation_id)}`}
            target="_blank"
            rel="noreferrer"
          >
            <ExternalLink size={14} /> Manage installation
          </a>
          {canManage && (
            <button
              type="button"
              className="inline-flex min-h-10 items-center gap-2 border border-rose-500/30 px-3 text-xs font-bold text-rose-500 transition hover:bg-rose-500/[0.08] disabled:opacity-50"
              disabled={disconnecting}
              onClick={() => {
                if (confirming) onDisconnect();
                setConfirming(!confirming);
              }}
            >
              <LogOut size={14} />
              {disconnecting ? "Disconnecting..." : confirming ? "Confirm disconnect" : "Disconnect"}
            </button>
          )}
        </div>
      )}
    </article>
  );
}

function Detail({
  label,
  value,
  icon,
}: {
  label: string;
  value: string;
  icon?: React.ReactNode;
}) {
  return (
    <div className="min-w-0">
      <dt className="text-[10px] font-semibold uppercase text-forge-muted">{label}</dt>
      <dd className="mt-1 flex items-center gap-1.5 truncate text-xs font-semibold text-forge-text">
        {icon}{value}
      </dd>
    </div>
  );
}