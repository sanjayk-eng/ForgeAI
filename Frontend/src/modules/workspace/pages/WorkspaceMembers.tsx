import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { UserPlus } from "lucide-react";
import { useAuth } from "../../auth/useAuth";
import { useToast } from "../../../shared/ui/useToast";
import {
  listInvites,
  listMembers,
  removeMember,
  updateMemberRole,
} from "../api";
import { InviteDialog } from "../components/InviteDialog";
import { WorkspaceInviteList } from "../components/WorkspaceInviteList";
import { WorkspaceMemberTable } from "../components/WorkspaceMemberTable";
import { useWorkspaceId } from "../hooks/useWorkspaceId";
import { useWorkspaceStore } from "../workspaceStore";

export function WorkspaceMembers() {
  const [view, setView] = useState<"members" | "invitations">("members");
  const workspaceId = useWorkspaceId();
  const { tokens } = useAuth();
  const toast = useToast();
  const queryClient = useQueryClient();
  const accessToken = tokens?.access_token ?? "";
  const inviteOpen = useWorkspaceStore((state) => state.inviteMemberOpen);
  const openInvite = useWorkspaceStore((state) => state.openInviteMember);
  const closeInvite = useWorkspaceStore((state) => state.closeInviteMember);
  const membersQuery = useQuery({
    queryKey: ["members", workspaceId],
    queryFn: () => listMembers(accessToken, workspaceId),
    enabled: Boolean(accessToken && workspaceId),
  });
  const invitesQuery = useQuery({
    queryKey: ["invites", workspaceId],
    queryFn: () => listInvites(accessToken, workspaceId),
    enabled: Boolean(accessToken && workspaceId),
  });
  const roleMutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      updateMemberRole(accessToken, workspaceId, userId, role),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["members", workspaceId] }),
    onError: () => toast.pushError("Could not update member role"),
  });
  const removeMutation = useMutation({
    mutationFn: (userId: string) =>
      removeMember(accessToken, workspaceId, userId),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["members", workspaceId] }),
    onError: () => toast.pushError("Could not remove member"),
  });
  const members = membersQuery.data ?? [];

  function closeInviteAndRefresh() {
    closeInvite();
    void queryClient.invalidateQueries({ queryKey: ["invites", workspaceId] });
  }

  return (
    <div className="animate-page-enter">
      <header className="mb-7 flex flex-col items-start justify-between gap-5 border-b border-white/[0.08] pb-6 sm:mb-8 sm:flex-row sm:items-end">
        <div>
          <span className="mb-3 block font-mono text-[10px] uppercase tracking-[.14em] text-forge-accent">
            Workspace / Access
          </span>
          <h1 className="m-0 text-3xl font-extrabold tracking-[-.05em] text-forge-text sm:text-[42px]">
            Access management
          </h1>
          <p className="mt-3 text-sm text-forge-muted">
            Manage current members and keep invitations moving separately.
          </p>
        </div>
        <button
          className="inline-flex items-center gap-2 rounded-md bg-forge-accent px-4 py-3 text-xs font-extrabold text-forge-bg transition hover:bg-[#d7ff82] focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
          onClick={openInvite}
        >
          <UserPlus size={16} />
          Invite member
        </button>
      </header>
      <div className="mb-4 flex w-full max-w-[420px] gap-1 border border-white/[0.09] bg-forge-panel/60 p-1">
        <button
          className={`flex flex-1 items-center justify-center gap-2 px-3 py-2.5 text-xs font-bold transition ${view === "members" ? "bg-forge-accent text-forge-bg" : "text-forge-muted hover:bg-white/[0.05] hover:text-forge-text"}`}
          onClick={() => setView("members")}
          aria-pressed={view === "members"}
        >
          Members
          <span
            className={`font-mono text-[10px] ${view === "members" ? "text-forge-bg/70" : "text-forge-muted"}`}
          >
            {members.length}
          </span>
        </button>
        <button
          className={`flex flex-1 items-center justify-center gap-2 px-3 py-2.5 text-xs font-bold transition ${view === "invitations" ? "bg-forge-accent text-forge-bg" : "text-forge-muted hover:bg-white/[0.05] hover:text-forge-text"}`}
          onClick={() => setView("invitations")}
          aria-pressed={view === "invitations"}
        >
          Invitations
          <span
            className={`font-mono text-[10px] ${view === "invitations" ? "text-forge-bg/70" : "text-forge-muted"}`}
          >
            {invitesQuery.data?.length ?? 0}
          </span>
        </button>
      </div>
      {view === "members" ? (
        <section className="border border-white/[0.09] bg-forge-panel/75">
          <div className="flex items-center justify-between border-b border-white/[0.08] p-5 sm:p-6">
            <div>
              <h2 className="m-0 text-base text-forge-text">Members</h2>
              <p className="mt-2 text-sm text-forge-muted">
                Everyone currently connected to this workspace.
              </p>
            </div>
            <span className="font-mono text-[10px] uppercase tracking-[.1em] text-forge-muted">
              {membersQuery.isFetching ? "Syncing" : "Synced"}
            </span>
          </div>
          <WorkspaceMemberTable
            members={members}
            loading={membersQuery.isLoading}
            error={membersQuery.isError}
            onRetry={() => void membersQuery.refetch()}
            onRoleChange={(userId, role) =>
              roleMutation.mutate({ userId, role })
            }
            onRemove={(userId) => removeMutation.mutate(userId)}
            roleUpdating={roleMutation.isPending}
            removing={removeMutation.isPending}
          />
        </section>
      ) : (
        <section className="border border-white/[0.09] bg-forge-panel/75">
          <div className="border-b border-white/[0.08] p-5 sm:p-6">
            <h2 className="m-0 text-base text-forge-text">Invitations</h2>
            <p className="mt-2 text-sm text-forge-muted">
              Track invitations sent to people who have not joined yet.
            </p>
          </div>
          <WorkspaceInviteList
            invites={invitesQuery.data ?? []}
            loading={invitesQuery.isLoading}
            error={invitesQuery.isError}
            onRetry={() => void invitesQuery.refetch()}
          />
        </section>
      )}
      {inviteOpen && (
        <InviteDialog
          accessToken={accessToken}
          workspaceId={workspaceId}
          onClose={closeInviteAndRefresh}
        />
      )}
    </div>
  );
}
