import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "../../../auth/useAuth";
import { useToast } from "../../../../shared/ui/useToast";
import { listWorkspaces, updateWorkspace } from "../../api/workspace.api";
import { useWorkspaceId } from "../../hooks/useWorkspaceId";

export function SettingsPage() {
  const id = useWorkspaceId();
  const { tokens } = useAuth();
  const toast = useToast();
  const client = useQueryClient();
  const query = useQuery({
    queryKey: ["workspaces"],
    queryFn: () => listWorkspaces(tokens?.access_token ?? ""),
    enabled: Boolean(tokens?.access_token),
  });
  const workspace = query.data?.find((item) => item.id === id);
  const [name, setName] = useState("");
  const value = name || workspace?.name || "";
  const mutation = useMutation({
    mutationFn: () =>
      updateWorkspace(tokens?.access_token ?? "", id, value.trim()),
    onSuccess: (updated) =>
      client.setQueryData(["workspaces"], (current: typeof query.data) =>
        current?.map((item) => (item.id === updated.id ? updated : item)),
      ),
    onError: () => toast.pushError("Could not save workspace settings"),
  });
  return (
    <div className="animate-page-enter">
      <div className="mb-7 sm:mb-10">
        <span className="mb-2 block font-mono text-[10px] uppercase tracking-[.12em] text-forge-accent">
          Workspace / Settings
        </span>
        <h1 className="m-0 text-3xl font-extrabold tracking-[-.045em] sm:text-[42px]">
          Workspace settings
        </h1>
        <p className="mt-3 text-sm text-forge-muted">
          Keep the workspace identity clear for everyone on the team.
        </p>
      </div>
      <section className="max-w-[680px] border border-white/[0.09] bg-forge-panel/75 p-7">
        <h2 className="m-0 text-base">General</h2>
        <p className="mt-2 mb-7 text-sm text-forge-muted">
          Update the name used across ForgeAI.
        </p>
        <label
          className="block font-mono text-[11px] text-forge-soft"
          htmlFor="settings-name"
        >
          Workspace name
        </label>
        <input
          id="settings-name"
          className="mt-2 mb-5 w-full max-w-[470px] rounded-md border border-white/[0.13] bg-[#101217] p-3 text-forge-text outline-none focus:border-forge-accent"
          value={value}
          onChange={(event) => setName(event.target.value)}
        />
        <br />
        <button
          className="rounded-md bg-forge-accent px-4 py-3 text-xs font-extrabold text-forge-bg disabled:cursor-not-allowed disabled:opacity-45"
          disabled={
            mutation.isPending ||
            value.trim().length < 2 ||
            value === workspace?.name
          }
          onClick={() => mutation.mutate()}
        >
          {mutation.isPending ? "Saving..." : "Save changes"}
        </button>
      </section>
    </div>
  );
}
