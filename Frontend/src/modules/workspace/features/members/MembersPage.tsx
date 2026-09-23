import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { UserPlus } from "lucide-react";
import { useAuth } from "../../../auth/useAuth";
import { useWorkspaceId } from "../../hooks/useWorkspaceId";
import { useWorkspaceStore } from "../../workspaceStore";
import { useMembersQuery } from "./useMembersQuery";
import { useInvitesQuery } from "../invites/useInvitesQuery";
import { WorkspaceMemberTable } from "../../components/WorkspaceMemberTable";
import { WorkspaceInviteList } from "../../components/WorkspaceInviteList";
import { InviteDialog } from "../../components/InviteDialog";

export function MembersPage() {
  const [view, setView] = useState<"members" | "invitations">("members");
  const workspaceId = useWorkspaceId();
  const { tokens } = useAuth();
  const queryClient = useQueryClient();
  const accessToken = tokens?.access_token ?? "";
  
  const inviteOpen = useWorkspaceStore((state) => state.inviteMemberOpen);
  const openInvite = useWorkspaceStore((state) => state.openInviteMember);
  const closeInvite = useWorkspaceStore((state) => state.closeInviteMember);

  const {
    members,
    isLoading: membersLoading,
    isError: membersError,
    refetch: refetchMembers,
    updateRole,
    removeMember,
    isUpdating,
  } = useMembersQuery(accessToken, workspaceId);

  const {
    invites,
    isLoading: invitesLoading,
    isError: invitesError,
    refetch: refetchInvites,
    revokeInvite,
    isRevoking,
  } = useInvitesQuery(accessToken, workspaceId);

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
            {invites.length}
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
              {membersLoading ? "Syncing" : "Synced"}
            </span>
          </div>
          <WorkspaceMemberTable
            members={members}
            loading={membersLoading}
            error={membersError}
            onRetry={() => void refetchMembers()}
            onRoleChange={updateRole}
            onRemove={removeMember}
            roleUpdating={isUpdating}
            removing={isUpdating}
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
            invites={invites}
            loading={invitesLoading}
            error={invitesError}
            onRetry={() => void refetchInvites()}
            onRevoke={revokeInvite}
            revoking={isRevoking}
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
