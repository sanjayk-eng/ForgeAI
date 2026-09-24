import { useState, type FormEvent } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { X } from "lucide-react";
import { useToast } from "../../../shared/ui/useToast";
import { createWorkspace } from "../api/workspace.api";
import type { Workspace } from "../types/workspace.types";

export function CreateWorkspaceDialog({
  accessToken,
  onClose,
  onCreated,
}: {
  accessToken: string;
  onClose: () => void;
  onCreated: (workspace: Workspace) => void;
}) {
  const [name, setName] = useState("");
  const toast = useToast();
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: () => createWorkspace(accessToken, name.trim()),
    onSuccess: (workspace) => {
      queryClient.setQueryData<Workspace[]>(["workspaces"], (current) => [
        ...(current ?? []),
        workspace,
      ]);
      onCreated(workspace);
    },
    onError: (error) =>
      toast.pushError(
        error instanceof Error ? error.message : "Could not create workspace",
      ),
  });
  function submit(event: FormEvent) {
    event.preventDefault();
    if (name.trim().length >= 2) mutation.mutate();
  }
  return (
    <div className="fixed inset-0 z-30 grid place-items-center bg-[var(--overlay)] p-5 backdrop-blur-md">
      <form
        className="w-full max-w-[450px] rounded-xl border border-forge-accent/20 bg-forge-card p-7 shadow-2xl"
        onSubmit={submit}
      >
        <div className="flex items-start justify-between gap-5">
          <div>
            <span className="mb-2 block font-mono text-[10px] uppercase tracking-[.12em] text-forge-accent">
              New workspace
            </span>
            <h2 className="m-0 text-[22px] font-bold tracking-[-.03em]">
              Create a workspace
            </h2>
          </div>
          <button
            type="button"
            className="grid size-9 place-items-center rounded-md text-forge-muted hover:bg-[var(--surface-hover)]"
            onClick={onClose}
            aria-label="Close dialog"
          >
            <X size={18} />
          </button>
        </div>
        <p className="mt-3 text-sm leading-6 text-forge-muted">
          A focused home for your team, agents, and work.
        </p>
        <label
          className="mt-6 block font-mono text-[11px] text-forge-soft"
          htmlFor="workspace-name"
        >
          Workspace name
        </label>
        <input
          id="workspace-name"
          className="mt-2 w-full rounded-md border border-[var(--border)] bg-[var(--input)] px-3 py-3 text-forge-text outline-none transition focus:border-forge-accent focus:ring-2 focus:ring-forge-accent/15"
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="e.g. Acme AI"
          autoFocus
        />
        <div className="mt-7 flex justify-end gap-2">
          <button
            type="button"
            className="rounded-md bg-[var(--surface-subtle)] px-4 py-2.5 text-xs font-extrabold text-forge-soft hover:bg-[var(--surface-hover)]"
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            className="rounded-md bg-forge-accent px-4 py-2.5 text-xs font-extrabold text-[var(--primary-foreground)] hover:bg-forge-accent-strong disabled:cursor-not-allowed disabled:opacity-45"
            disabled={mutation.isPending || name.trim().length < 2}
          >
            {mutation.isPending ? "Creating..." : "Create workspace"}
          </button>
        </div>
      </form>
    </div>
  );
}
