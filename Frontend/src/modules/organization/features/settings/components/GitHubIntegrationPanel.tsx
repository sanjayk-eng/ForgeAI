import { GitBranch, ShieldCheck } from "lucide-react";
import { GitHubConnectedDetails } from "./GitHubConnectedDetails";
import type { OrganizationIntegration } from "../../../api/integrations.api";

export function GitHubIntegrationPanel({
  integrations,
  canManage,
  connecting,
  disconnectingId,
  onConnect,
  onDisconnect,
}: {
  integrations: OrganizationIntegration[];
  canManage: boolean;
  connecting: boolean;
  disconnectingId: string | null;
  onConnect: () => void;
  onDisconnect: (integrationId: string) => void;
}) {
  const activeIntegrations = integrations.filter((item) => item.status === "ACTIVE");

  return (
    <section className="min-w-0 p-4 sm:p-6">
      <header className="flex flex-wrap items-start justify-between gap-4 border-b border-[var(--border)] pb-5">
        <div className="flex min-w-0 items-start gap-3">
          <GitBranch className="mt-0.5 shrink-0 text-forge-text" size={30} />
          <div>
            <h2 className="m-0 text-base font-bold text-forge-text">GitHub</h2>
            <p className="mt-1 max-w-[520px] text-xs leading-5 text-forge-muted">
              Connect an authorized GitHub App installation to this ForgeAI organization.
            </p>
          </div>
        </div>
        {activeIntegrations.length > 0 ? (
          <span className="inline-flex items-center gap-2 border border-emerald-500/20 bg-emerald-500/[0.08] px-2.5 py-1 text-[10px] font-bold text-emerald-500">
            <span className="size-1.5 rounded-full bg-emerald-500" /> Connected
          </span>
        ) : integrations.length > 0 ? (
          <span className="inline-flex items-center gap-2 border border-amber-500/20 bg-amber-500/[0.08] px-2.5 py-1 text-[10px] font-bold text-amber-500">
            Installation needs attention
          </span>
        ) : null}
      </header>

      {activeIntegrations.length === 0 && (
        <div className="mt-5 border border-dashed border-[var(--border)] bg-[var(--surface-subtle)] p-5 sm:p-6">
          <div className="flex flex-col items-start justify-between gap-5 sm:flex-row sm:items-center">
            <div className="flex items-start gap-3">
              <span className="grid size-9 shrink-0 place-items-center border border-[var(--border)] bg-forge-panel text-forge-text"><GitBranch size={18} /></span>
              <div>
                <h3 className="m-0 text-sm font-bold text-forge-text">
                  {integrations.length > 0 ? "GitHub installation is inactive" : "GitHub is not connected"}
                </h3>
                <p className="mt-1 max-w-[440px] text-xs leading-5 text-forge-muted">
                  {integrations.length > 0
                    ? "The installation is linked to this organization but is not active. Check its status below or authorize an installation again."
                    : "Install the ForgeAI GitHub App for a GitHub personal account or organization."}
                </p>
              </div>
            </div>
            {canManage && (
              <button
                type="button"
                className="inline-flex min-h-10 shrink-0 items-center gap-2 bg-forge-accent px-4 text-xs font-extrabold text-[var(--primary-foreground)] transition hover:bg-forge-accent-strong disabled:cursor-wait disabled:opacity-50"
                disabled={connecting}
                onClick={onConnect}
              >
                <GitBranch size={15} /> {connecting ? "Opening GitHub..." : "Connect GitHub"}
              </button>
            )}
          </div>
          {!canManage && (
            <p className="mt-4 border-t border-[var(--border)] pt-3 text-xs text-forge-muted">
              An organization owner or admin can connect GitHub.
            </p>
          )}
        </div>
      )}

      {integrations.length > 0 && (
        <div className="mt-5 grid gap-3">
          {integrations.map((integration) => (
            <GitHubConnectedDetails
              key={integration.id}
              integration={integration}
              canManage={canManage}
              disconnecting={disconnectingId === integration.id}
              onDisconnect={() => onDisconnect(integration.id)}
            />
          ))}
        </div>
      )}

      <div className="mt-6 grid gap-4 border-t border-[var(--border)] pt-5 sm:grid-cols-[1fr_auto] sm:items-start">
        <div className="flex items-start gap-3">
          <GitBranch className="mt-0.5 shrink-0 text-forge-text" size={20} />
          <div>
            <h3 className="m-0 text-xs font-bold text-forge-text">About GitHub integration</h3>
            <p className="mt-2 max-w-[480px] text-xs leading-5 text-forge-muted">
              This associates a GitHub App installation with this ForgeAI organization. Project contents remain in ForgeAI until explicitly pushed; GitHub repositories are never imported or synced back.
            </p>
          </div>
        </div>
        <div className="flex max-w-[250px] items-start gap-2 border border-[var(--border)] bg-[var(--surface-subtle)] p-3 text-xs text-forge-muted">
          <ShieldCheck className="shrink-0 text-forge-accent" size={15} />
          <span>Only organization owners and admins can manage the installation.</span>
        </div>
      </div>
    </section>
  );
}