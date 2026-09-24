import { useQuery } from "@tanstack/react-query";
import { Outlet, useSearchParams } from "react-router-dom";
import { ArrowRight, ShieldCheck } from "lucide-react";
import { useAuth } from "../auth/useAuth";
import { listWorkspaces } from "./api/workspace.api";
import type { Workspace } from "./types/workspace.types";
import { CreateWorkspaceDialog } from "./components/CreateWorkspaceDialog";
import { WorkspaceHeader } from "./components/WorkspaceHeader";
import { WorkspaceSidebar } from "./components/WorkspaceSidebar";
import { useWorkspaceStore } from "./workspaceStore";

export function WorkspaceLayout() {
  const { user, tokens, signOut } = useAuth();
  const [params, setParams] = useSearchParams();
  const mobileOpen = useWorkspaceStore((state) => state.mobileNavigationOpen);
  const createOpen = useWorkspaceStore((state) => state.createWorkspaceOpen);
  const openMobileNavigation = useWorkspaceStore(
    (state) => state.openMobileNavigation,
  );
  const closeMobileNavigation = useWorkspaceStore(
    (state) => state.closeMobileNavigation,
  );
  const openCreateWorkspace = useWorkspaceStore(
    (state) => state.openCreateWorkspace,
  );
  const closeCreateWorkspace = useWorkspaceStore(
    (state) => state.closeCreateWorkspace,
  );
  const accessToken = tokens?.access_token ?? "";
  const query = useQuery({
    queryKey: ["workspaces"],
    queryFn: () => listWorkspaces(accessToken),
    enabled: Boolean(accessToken),
  });
  const workspaces = query.data ?? [];
  const selected =
    workspaces.find((item) => item.id === params.get("workspace")) ??
    workspaces[0];

  function selectWorkspace(workspace: Workspace) {
    setParams({ workspace: workspace.id });
  }

  return (
    <div className="min-h-screen bg-forge-bg text-forge-text">
      <WorkspaceHeader
        userName={user?.name}
        workspaces={workspaces}
        selected={selected}
        onSelect={selectWorkspace}
        onMenu={openMobileNavigation}
        onCreateWorkspace={openCreateWorkspace}
        accessToken={accessToken}
      />
      <div className="flex h-[calc(100vh-72px)] min-h-0 overflow-hidden">
        <WorkspaceSidebar
          open={mobileOpen}
          hasWorkspace={Boolean(selected)}
          workspaceId={selected?.id}
          onCreateWorkspace={openCreateWorkspace}
          onClose={closeMobileNavigation}
          onSignOut={signOut}
        />
        <main className="min-h-0 min-w-0 flex-1 overflow-y-auto">
          <div className="mx-auto w-[calc(100%-32px)] max-w-[1180px] py-8 sm:w-[calc(100%-64px)] sm:py-11 lg:w-[calc(100%-96px)]">
            {query.isLoading ? (
              <LoadingState />
            ) : selected ? (
              <Outlet />
            ) : (
              <EmptyState onCreate={openCreateWorkspace} />
            )}
          </div>
        </main>
      </div>
      {createOpen && (
        <CreateWorkspaceDialog
          accessToken={accessToken}
          onClose={closeCreateWorkspace}
          onCreated={(workspace) => {
            selectWorkspace(workspace);
            closeCreateWorkspace();
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
      Loading your workspace
    </div>
  );
}

function EmptyState({ onCreate }: { onCreate: () => void }) {
  return (
    <div className="mx-auto grid min-h-[480px] max-w-[760px] place-content-center">
      <section className="border border-white/[0.09] bg-forge-panel/80 p-7 shadow-[0_24px_70px_rgba(0,0,0,.18)] sm:p-10">
        <div className="flex flex-col gap-8 sm:flex-row sm:items-start sm:justify-between">
          <div className="max-w-[460px]">
            <span className="mb-3 block font-mono text-[10px] uppercase tracking-[.14em] text-forge-accent">
              Workspace setup
            </span>
            <div className="mb-5 flex size-11 items-center justify-center border border-forge-accent/25 bg-forge-accent/[0.08] text-forge-accent">
              <ShieldCheck size={22} />
            </div>
            <h1 className="m-0 text-3xl font-extrabold tracking-[-.045em] text-[#f1f3f5] sm:text-4xl">
              Create your first workspace
            </h1>
            <p className="mt-4 max-w-[420px] text-sm leading-6 text-forge-muted">
              Start a focused home for your team, agents, repositories, and
              conversations.
            </p>
            <button
              className="group mt-7 inline-flex items-center gap-2 rounded-md bg-forge-accent px-4 py-3 text-xs font-extrabold text-forge-bg transition hover:bg-[#d7ff82] focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
              onClick={onCreate}
            >
              Create workspace
              <ArrowRight
                size={16}
                className="transition-transform group-hover:translate-x-1"
              />
            </button>
          </div>
          <div className="grid min-w-[190px] gap-4 border-t border-white/[0.08] pt-5 text-xs text-forge-muted sm:border-l sm:border-t-0 sm:pl-6 sm:pt-0">
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
