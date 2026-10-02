import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { Outlet, useLocation, useSearchParams } from "react-router-dom";
import { ArrowRight, ShieldCheck } from "lucide-react";
import { useAuth } from "../auth/useAuth";
import { listOrganizations } from "./api/organization.api";
import type { Organization } from "./types/organization.types";
import { CreateOrganizationDialog } from "./components/CreateOrganizationDialog";
import { OrganizationHeader } from "./components/OrganizationHeader";
import { OrganizationSidebar } from "./components/OrganizationSidebar";
import { useOrganizationStore } from "./organizationStore";

export function OrganizationLayout() {
  const { pathname } = useLocation();
  const isProjectDetail = /^\/organizations\/[^/]+\/projects\/[^/]+$/.test(pathname);
  const { user, tokens, signOut } = useAuth();
  const [params, setParams] = useSearchParams();
  const mobileOpen = useOrganizationStore((state) => state.mobileNavigationOpen);
  const createOpen = useOrganizationStore((state) => state.createOrganizationOpen);
  const openMobileNavigation = useOrganizationStore(
    (state) => state.openMobileNavigation,
  );
  const closeMobileNavigation = useOrganizationStore(
    (state) => state.closeMobileNavigation,
  );
  const openCreateOrganization = useOrganizationStore(
    (state) => state.openCreateOrganization,
  );
  const closeCreateOrganization = useOrganizationStore(
    (state) => state.closeCreateOrganization,
  );
  const accessToken = tokens?.access_token ?? "";
  const query = useQuery({
    queryKey: ["organizations"],
    queryFn: () => listOrganizations(accessToken),
    enabled: Boolean(accessToken),
  });
  const organizations = query.data ?? [];
  const selected =
    organizations.find((item) => item.id === params.get("organization")) ??
    organizations[0];

  useEffect(() => {
    if (query.isLoading || query.isError || !selected) return;
    if (params.get("organization") !== selected.id) {
      setParams({ organization: selected.id }, { replace: true });
    }
  }, [params, query.isError, query.isLoading, selected, setParams]);

  function selectOrganization(organization: Organization) {
    setParams({ organization: organization.id });
  }

  return (
    <div className="min-h-screen bg-forge-bg text-forge-text">
      <OrganizationHeader
        userName={user?.name}
        organizations={organizations}
        selected={selected}
        onSelect={selectOrganization}
        onMenu={openMobileNavigation}
        onCreateOrganization={openCreateOrganization}
        accessToken={accessToken}
      />
      <div className="flex h-[calc(100vh-72px)] min-h-0 overflow-hidden">
        <OrganizationSidebar
          open={mobileOpen}
          hasOrganization={Boolean(selected)}
          organizationId={selected?.id}
          onCreateOrganization={openCreateOrganization}
          onClose={closeMobileNavigation}
          onSignOut={signOut}
        />
        <main className={`min-h-0 min-w-0 flex-1 ${isProjectDetail ? "overflow-hidden" : "overflow-y-auto"}`}>
          <div className={isProjectDetail ? "h-full min-h-0 w-full max-w-none" : "mx-auto w-[calc(100%-32px)] max-w-[1180px] py-8 sm:w-[calc(100%-64px)] sm:py-11 lg:w-[calc(100%-96px)]"}>
            {query.isLoading ? (
              <LoadingState />
            ) : selected ? (
              <Outlet />
            ) : (
              <EmptyState onCreate={openCreateOrganization} />
            )}
          </div>
        </main>
      </div>
      {createOpen && (
        <CreateOrganizationDialog
          accessToken={accessToken}
          onClose={closeCreateOrganization}
          onCreated={(organization: Organization) => {
            selectOrganization(organization);
            closeCreateOrganization();
          }}
        />
      )}
    </div>
  );
}

function LoadingState() {
  return (
    <div className="grid min-h-[380px] place-content-center justify-items-center gap-3 text-sm text-forge-muted">
      <span className="size-2.5 animate-pulse rounded-full bg-forge-accent" />
      Loading your organization
    </div>
  );
}

function EmptyState({ onCreate }: { onCreate: () => void }) {
  return (
    <div className="mx-auto grid min-h-[480px] max-w-[760px] place-content-center">
      <section className="border border-[var(--border)] bg-forge-panel/80 p-7 shadow-[0_24px_70px_rgba(0,0,0,.12)] sm:p-10">
        <div className="flex flex-col gap-8 sm:flex-row sm:items-start sm:justify-between">
          <div className="max-w-[460px]">
            <span className="mb-3 block font-mono text-[10px] uppercase tracking-[.14em] text-forge-accent">
              Organization setup
            </span>
            <div className="mb-5 flex size-11 items-center justify-center border border-forge-accent/25 bg-forge-accent/[0.08] text-forge-accent">
              <ShieldCheck size={22} />
            </div>
            <h1 className="m-0 text-3xl font-extrabold tracking-[-.045em] text-forge-text sm:text-4xl">
              Create your first organization
            </h1>
            <p className="mt-4 max-w-[420px] text-sm leading-6 text-forge-muted">
              Start a focused home for your team, agents, projects, and
              conversations.
            </p>
            <button
              className="group mt-7 inline-flex items-center gap-2 rounded-md bg-forge-accent px-4 py-3 text-xs font-extrabold text-forge-bg transition hover:bg-forge-accent-strong focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
              onClick={onCreate}
            >
              Create organization
              <ArrowRight
                size={16}
                className="transition-transform group-hover:translate-x-1"
              />
            </button>
          </div>
          <div className="grid min-w-[190px] gap-4 border-t border-[var(--border)] pt-5 text-xs text-forge-muted sm:border-l sm:border-t-0 sm:pl-6 sm:pt-0">
            <div>
              <strong className="mb-1 block text-forge-soft">01</strong>Create a
              shared home
            </div>
            <div>
              <strong className="mb-1 block text-forge-soft">02</strong>Invite
              your team
            </div>
            <div>
              <strong className="mb-1 block text-forge-soft">03</strong>Build
              together
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
