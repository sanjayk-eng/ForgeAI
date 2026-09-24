import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { ChevronRight, UserPlus } from "lucide-react";
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
    <div className="animate-page-enter mx-auto max-w-[1040px]">
      <header className="mb-7 flex flex-col items-start justify-between gap-5 border-b border-[var(--border)] pb-7 sm:mb-8 sm:flex-row sm:items-end">
        <div className="min-w-0">
          <div className="mb-4 flex items-center gap-2 text-xs font-semibold text-forge-muted">
            <span>Workspace</span><ChevronRight size={13} /><span className="text-forge-soft">Members</span>
          </div>
          <h1 className="m-0 text-3xl font-extrabold tracking-[-.035em] text-forge-text sm:text-[36px]">
            Access management
          </h1>
          <p className="mt-3 max-w-[560px] text-sm leading-6 text-forge-muted">
            Manage current members and keep invitations moving separately.
          </p>
        </div>
        <button
          className="inline-flex h-11 items-center gap-2 rounded-lg bg-forge-accent px-4 text-xs font-extrabold text-[var(--primary-foreground)] shadow-[0_8px_24px_rgba(198,243,106,.12)] transition hover:bg-forge-accent-strong active:translate-y-px disabled:cursor-not-allowed disabled:opacity-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-forge-accent"
          onClick={openInvite}
        >
          <UserPlus size={16} />
          Invite member
        </button>
      </header>

      <div className="mb-5 flex w-full max-w-[420px] gap-1 rounded-lg border border-[var(--border)] bg-forge-panel/60 p-1" role="tablist" aria-label="Access management views">
        <button
          className={`flex flex-1 items-center justify-center gap-2 rounded-md px-3 py-2.5 text-xs font-bold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent ${view === "members" ? "bg-forge-accent text-[var(--primary-foreground)] shadow-sm" : "text-forge-muted hover:bg-white/[0.05] hover:text-forge-text"}`}
          onClick={() => setView("members")}
          aria-pressed={view === "members"}
          role="tab"
        >
          Members
          <span
            className={`rounded-full px-1.5 py-0.5 font-mono text-[10px] ${view === "members" ? "bg-forge-bg/15 text-forge-bg" : "bg-white/[0.06] text-forge-muted"}`}
          >
            {members.length}
          </span>
        </button>
        <button
          className={`flex flex-1 items-center justify-center gap-2 rounded-md px-3 py-2.5 text-xs font-bold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent ${view === "invitations" ? "bg-forge-accent text-[var(--primary-foreground)] shadow-sm" : "text-forge-muted hover:bg-white/[0.05] hover:text-forge-text"}`}
          onClick={() => setView("invitations")}
          aria-pressed={view === "invitations"}
          role="tab"
        >
          Invitations
          <span
            className={`rounded-full px-1.5 py-0.5 font-mono text-[10px] ${view === "invitations" ? "bg-forge-bg/15 text-forge-bg" : "bg-white/[0.06] text-forge-muted"}`}
          >
            {invites.length}
          </span>
        </button>
      </div>

      {view === "members" ? (
        <section className="overflow-hidden rounded-xl border border-[var(--border)] bg-forge-panel/75 shadow-[0_18px_50px_rgba(0,0,0,.08)]">
          <div className="flex items-center justify-between border-b border-[var(--border)] p-5 sm:p-6">
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
        <section className="overflow-hidden rounded-xl border border-[var(--border)] bg-forge-panel/75 shadow-[0_18px_50px_rgba(0,0,0,.08)]">
          <div className="border-b border-[var(--border)] p-5 sm:p-6">
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
