import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { MoreHorizontal, UserPlus, Users } from "lucide-react";
import { useAuth } from "../../auth/useAuth";
import { useToast } from "../../../shared/ui/useToast";
import { listMembers, removeMember, updateMemberRole } from "../api";
import { useWorkspaceId } from "../hooks/useWorkspaceId";
import { InviteDialog } from "../components/InviteDialog";
import { useWorkspaceStore } from "../workspaceStore";

export function WorkspaceMembers() {
  const id = useWorkspaceId();
  const { tokens } = useAuth();
  const toast = useToast();
  const client = useQueryClient();
  const inviteOpen = useWorkspaceStore((state) => state.inviteMemberOpen);
  const openInvite = useWorkspaceStore((state) => state.openInviteMember);
  const closeInvite = useWorkspaceStore((state) => state.closeInviteMember);
  const query = useQuery({
    queryKey: ["members", id],
    queryFn: () => listMembers(tokens?.access_token ?? "", id),
    enabled: Boolean(tokens?.access_token && id),
  });
  const members = query.data ?? [];
  const role = useMutation({
    mutationFn: ({ userId, value }: { userId: string; value: string }) =>
      updateMemberRole(tokens?.access_token ?? "", id, userId, value),
    onSuccess: () => client.invalidateQueries({ queryKey: ["members", id] }),
    onError: () => toast.pushError("Could not update member role"),
  });
  const remove = useMutation({
    mutationFn: (userId: string) =>
      removeMember(tokens?.access_token ?? "", id, userId),
    onSuccess: () => client.invalidateQueries({ queryKey: ["members", id] }),
    onError: () => toast.pushError("Could not remove member"),
  });
  return (
    <div className="animate-page-enter">
      <div className="mb-7 flex flex-col items-start justify-between gap-5 sm:mb-10 sm:flex-row sm:items-end">
        <div>
          <span className="mb-2 block font-mono text-[10px] uppercase tracking-[.12em] text-forge-accent">
            Workspace / Members
          </span>
          <h1 className="m-0 text-3xl font-extrabold tracking-[-.045em] sm:text-[42px]">
            People with access
          </h1>
          <p className="mt-3 text-sm text-forge-muted">
            Invite collaborators and shape their workspace role.
          </p>
        </div>
        <button
          className="flex items-center gap-2 rounded-md bg-forge-accent px-4 py-3 text-xs font-extrabold text-forge-bg"
          onClick={openInvite}
        >
          <UserPlus size={16} />
          Invite member
        </button>
      </div>
      <section className="border border-white/[0.09] bg-forge-panel/75">
        <div className="flex items-center justify-between border-b border-white/[0.08] p-6">
          <div>
            <h2 className="m-0 text-base">
              Members{" "}
              <span className="ml-1 inline-grid min-w-5 place-items-center rounded border border-forge-accent/25 px-1.5 py-1 font-mono text-[11px] text-forge-accent">
                {members.length}
              </span>
            </h2>
            <p className="mt-2 text-sm text-forge-muted">
              Everyone currently connected to this workspace.
            </p>
          </div>
          <button className="text-forge-muted" aria-label="Member actions">
            <MoreHorizontal size={18} />
          </button>
        </div>
        {query.isLoading ? (
          <Empty text="Loading members..." />
        ) : members.length === 0 ? (
          <Empty text="No members found yet." icon />
        ) : (
          <div>
            {members.map((member) => (
              <div
                key={member.id}
                className="grid grid-cols-[1fr_auto] items-center gap-3 border-b border-white/[0.06] p-4 sm:grid-cols-[minmax(220px,1.8fr)_150px_130px_75px] sm:px-6"
              >
                <div className="flex min-w-0 items-center gap-3">
                  <span className="grid size-8 shrink-0 place-items-center rounded-full bg-[#8ed4ff] font-mono text-xs text-forge-bg">
                    {member.user_id.slice(0, 2).toUpperCase()}
                  </span>
                  <div className="min-w-0">
                    <strong className="block truncate text-[13px]">
                      {member.user_id}
                    </strong>
                    <small className="block truncate font-mono text-[10px] text-forge-muted">
                      {member.user_id}
                    </small>
                  </div>
                </div>
                <select
                  className="rounded border border-white/[0.11] bg-forge-card p-2 font-mono text-[11px] text-forge-soft"
                  value={member.role}
                  disabled={member.role === "OWNER"}
                  onChange={(event) =>
                    role.mutate({
                      userId: member.user_id,
                      value: event.target.value,
                    })
                  }
                >
                  <option>OWNER</option>
                  <option>ADMIN</option>
                  <option>MEMBER</option>
                </select>
                <span className="hidden font-mono text-[10px] text-forge-muted sm:block">
                  {new Date(member.created_at).toLocaleDateString()}
                </span>
                <button
                  className="col-start-2 row-start-1 text-[11px] text-forge-muted hover:text-[#ff8e7a] disabled:opacity-30 sm:col-auto sm:row-auto"
                  disabled={member.role === "OWNER"}
                  onClick={() => remove.mutate(member.user_id)}
                >
                  Remove
                </button>
              </div>
            ))}
          </div>
        )}
      </section>
      {inviteOpen && (
        <InviteDialog
          accessToken={tokens?.access_token ?? ""}
          workspaceId={id}
          onClose={closeInvite}
        />
      )}
    </div>
  );
}

function Empty({ text, icon = false }: { text: string; icon?: boolean }) {
  return (
    <div className="grid min-h-[170px] place-content-center justify-items-center gap-3 text-sm text-forge-muted">
      {icon && <Users size={24} />}
      {text}
    </div>
  );
}
