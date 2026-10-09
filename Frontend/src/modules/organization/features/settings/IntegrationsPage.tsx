import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronRight } from "lucide-react";
import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useAuth } from "../../../auth/useAuth";
import { useOrganizationId } from "../../hooks/useOrganizationId";
import { listMembers } from "../../api/members.api";
import {
  disconnectGitHubIntegration,
  listOrganizationIntegrations,
  startGitHubInstallation,
} from "../../api/integrations.api";
import { IntegrationProviderList } from "./components/IntegrationProviderList";
import { GitHubIntegrationPanel } from "./components/GitHubIntegrationPanel";

export function IntegrationsPage() {
  const organizationId = useOrganizationId();
  const { tokens, user } = useAuth();
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const [connectionNotice, setConnectionNotice] = useState(false);
  const accessToken = tokens?.access_token ?? "";

  const integrationsQuery = useQuery({
    queryKey: ["organization-integrations", organizationId],
    queryFn: () => listOrganizationIntegrations(accessToken, organizationId),
    enabled: Boolean(accessToken && organizationId),
  });
  const roleQuery = useQuery({
    queryKey: ["organization-current-role", organizationId, user?.id],
    queryFn: () => listMembers(accessToken, organizationId, {
      page: 1,
      perPage: 100,
      search: user?.email,
    }),
    enabled: Boolean(accessToken && organizationId && user?.id),
  });
  const canManage = roleQuery.data?.items.some(
    (member) => member.user.id === user?.id && (member.role === "OWNER" || member.role === "ADMIN"),
  ) ?? false;

  useEffect(() => {
    if (searchParams.get("github_integration") !== "connected") return;

    setConnectionNotice(true);
    void queryClient.invalidateQueries({
      queryKey: ["organization-integrations", organizationId],
    });

    const nextSearchParams = new URLSearchParams(searchParams);
    nextSearchParams.delete("github_integration");
    setSearchParams(nextSearchParams, { replace: true });
  }, [organizationId, queryClient, searchParams, setSearchParams]);

  const connectMutation = useMutation({
    mutationFn: () => startGitHubInstallation(accessToken, organizationId),
    onSuccess: (installUrl) => window.location.assign(installUrl),
  });
  const disconnectMutation = useMutation({
    mutationFn: (integrationId: string) =>
      disconnectGitHubIntegration(accessToken, organizationId, integrationId),
    onSuccess: () => queryClient.invalidateQueries({
      queryKey: ["organization-integrations", organizationId],
    }),
  });

  return (
    <div className="animate-page-enter mx-auto max-w-[1120px]">
      <header className="mb-6 border-b border-[var(--border)] pb-5">
        <div className="mb-3 flex items-center gap-2 text-xs font-semibold text-forge-muted">
          <span>Organization Settings</span><ChevronRight size={13} /><span className="text-forge-soft">Integrations</span>
        </div>
        <h1 className="m-0 text-2xl font-extrabold text-forge-text">Integrations</h1>
        <p className="mt-2 text-xs text-forge-muted">Connect the tools your organization uses.</p>
      </header>

      {connectionNotice && (
        <p className="mb-4 border border-emerald-500/20 bg-emerald-500/[0.08] px-4 py-3 text-xs font-semibold text-emerald-500" role="status">
          GitHub was authorized. Refreshing the connection for this organization.
        </p>
      )}

      <div className="grid overflow-hidden border border-[var(--border)] bg-forge-panel/60 md:grid-cols-[220px_minmax(0,1fr)]">
        <IntegrationProviderList />
        <div className="min-w-0">
          {integrationsQuery.isLoading ? (
            <div className="grid min-h-[360px] place-content-center text-sm text-forge-muted" role="status">Loading integrations...</div>
          ) : integrationsQuery.isError ? (
            <div className="grid min-h-[360px] place-content-center justify-items-center gap-3 p-6 text-center" role="alert">
              <p className="m-0 text-sm text-forge-muted">Could not load organization integrations.</p>
              <button className="text-xs font-bold text-forge-accent" onClick={() => void integrationsQuery.refetch()}>Try again</button>
            </div>
          ) : (
            <GitHubIntegrationPanel
              integrations={integrationsQuery.data ?? []}
              canManage={canManage}
              connecting={connectMutation.isPending}
              disconnectingId={disconnectMutation.variables ?? null}
              onConnect={() => connectMutation.mutate()}
              onDisconnect={(integrationId) => disconnectMutation.mutate(integrationId)}
            />
          )}
        </div>
      </div>
      {connectMutation.isError && <p className="mt-3 text-xs text-rose-500" role="alert">Could not start GitHub installation. Try again.</p>}
      {disconnectMutation.isError && <p className="mt-3 text-xs text-rose-500" role="alert">Could not disconnect GitHub. Verify the installation is still available and try again.</p>}
    </div>
  );
}