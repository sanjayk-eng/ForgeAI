import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AlertCircle, CheckCircle2, GitBranch, Loader2, RotateCcw, Send } from "lucide-react";
import { useMemo, useState, type FormEvent, type ReactNode } from "react";
import { getGitDiff, getGitStatus, pushGitChanges, commitGitChanges, revertGitChanges } from "../../../api/git.api";
import { GitDiffPreview } from "./GitDiffPreview";
import {
  buildCommitFileList,
  getGitSelectionSummary,
  nextGitSelection,
  resetGitSelection,
  type GitSelectionState,
} from "./gitPanelState";

export function GitPanel({
  accessToken,
  sandboxId,
  sandboxStatus,
  projectName,
}: {
  accessToken: string | null;
  sandboxId: string | null;
  sandboxStatus: string | null;
  projectName: string;
}) {
  const [message, setMessage] = useState("Initial commit");
  const [error, setError] = useState<string | null>(null);
  const [fileSelection, setFileSelection] = useState<GitSelectionState>(() => ({
    pathsKey: "",
    excluded: new Set(),
  }));
  const queryClient = useQueryClient();
  const sandboxReady = sandboxStatus === "RUNNING";

  const statusQuery = useQuery({
    queryKey: ["git-status", sandboxId],
    queryFn: () => {
      if (!accessToken || !sandboxId) throw new Error("Missing sandbox context");
      return getGitStatus(accessToken, sandboxId);
    },
    enabled: Boolean(accessToken && sandboxId && sandboxReady),
    staleTime: 15_000,
    retry: 3,
    retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 5000),
  });

  const diffQuery = useQuery({
    queryKey: ["git-diff", sandboxId],
    queryFn: () => {
      if (!accessToken || !sandboxId) throw new Error("Missing sandbox context");
      return getGitDiff(accessToken, sandboxId);
    },
    enabled: Boolean(accessToken && sandboxId && sandboxReady),
    staleTime: 15_000,
    retry: 3,
    retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 5000),
  });

  const status = sandboxReady ? statusQuery.data?.status : undefined;
  const files = diffQuery.data?.files ?? [];
  const changedPaths = useMemo(() => [...new Set([
    ...(status?.staged ?? []),
    ...(status?.modified ?? []),
    ...(status?.untracked ?? []),
    ...files.map((file) => file.path),
  ])].sort(), [files, status]);
  const selectionSummary = useMemo(
    () => getGitSelectionSummary(changedPaths, fileSelection),
    [changedPaths, fileSelection],
  );
  const selectedPaths = selectionSummary.selectedPaths;

  const commitMutation = useMutation({
    mutationFn: (paths: string[]) => {
      if (!accessToken || !sandboxId) throw new Error("Missing sandbox context");
      return commitGitChanges(accessToken, sandboxId, message.trim(), buildCommitFileList(paths, files));
    },
    onSuccess: () => {
      setError(null);
      void statusQuery.refetch();
      void diffQuery.refetch();
      if (sandboxId) void queryClient.invalidateQueries({ queryKey: ["git-file-diff", sandboxId] });
    },
    onError: (submitError) => {
      setError(submitError instanceof Error ? submitError.message : "Commit failed");
    },
  });

  const revertMutation = useMutation({
    mutationFn: async () => {
      if (!accessToken || !sandboxId) throw new Error("Missing sandbox context");
      return revertGitChanges(accessToken, sandboxId, selectedPaths);
    },
    onSuccess: () => {
      setError(null);
      setMessage("Initial commit");
    },
    onSettled: async () => {
      await Promise.all([statusQuery.refetch(), diffQuery.refetch()]);
      if (sandboxId) await queryClient.invalidateQueries({ queryKey: ["git-file-diff", sandboxId] });
    },
    onError: (submitError) => {
      setError(submitError instanceof Error ? submitError.message : "Revert failed");
    },
  });

  const pushMutation = useMutation({
    mutationFn: async () => {
      if (!accessToken || !sandboxId) throw new Error("Missing sandbox context");
      return pushGitChanges(accessToken, sandboxId);
    },
    onSuccess: () => {
      setError(null);
    },
    onSettled: async () => {
      await Promise.all([statusQuery.refetch(), diffQuery.refetch()]);
      if (sandboxId) await queryClient.invalidateQueries({ queryKey: ["git-file-diff", sandboxId] });
    },
    onError: (submitError) => {
      setError(submitError instanceof Error ? submitError.message : "Push failed");
    },
  });

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!message.trim()) {
      setError("Commit message is required");
      return;
    }
    if (selectedPaths.length === 0) {
      setError("Select at least one file to commit");
      return;
    }
    commitMutation.mutate(selectedPaths);
  }

  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col gap-3 p-3">
      <div className="max-h-[44%] shrink-0 space-y-4 overflow-y-auto pr-1">
      <div>
        <div className="flex items-center gap-2 text-xs font-semibold text-forge-text">
          <GitBranch size={14} className="text-forge-accent" />
          Git workspace
        </div>
        <p className="mt-2 text-xs leading-5 text-forge-muted">
          Track changes in {projectName}. Git is separate from the file browser and operates in the sandbox /workspace.
        </p>
      </div>

      {statusQuery.isLoading && (
        <div className="flex items-center gap-2 text-xs text-forge-muted"><Loader2 size={13} className="animate-spin" /> Checking git status</div>
      )}
      {!sandboxReady && (
        <StatusNotice>
          {sandboxStatus
            ? `Git becomes available when the sandbox is running (currently ${sandboxStatus.toLowerCase()}).`
            : "Waiting for sandbox status before checking Git."}
        </StatusNotice>
      )}
      {sandboxReady && statusQuery.isError && (
        <StatusNotice>
          <div className="flex items-start justify-between gap-3">
            <span>{statusQuery.error instanceof Error ? statusQuery.error.message : "Could not load Git status."}</span>
            <button
              type="button"
              className="shrink-0 underline underline-offset-2 disabled:opacity-50"
              disabled={statusQuery.isFetching || diffQuery.isFetching}
              onClick={() => {
                void statusQuery.refetch();
                void diffQuery.refetch();
              }}
            >
              Retry
            </button>
          </div>
        </StatusNotice>
      )}
      {status && (
        <div className="space-y-2 border border-[var(--border)] bg-forge-bg p-3">
          <div className="flex items-center justify-between gap-3">
            <span className="text-[10px] uppercase tracking-[0.12em] text-forge-muted">Branch</span>
            <span className="font-mono text-[11px] text-forge-text">{status.branch || "(detached)"}</span>
          </div>
          <div className="flex items-center justify-between gap-3">
            <span className="text-[10px] uppercase tracking-[0.12em] text-forge-muted">Remote</span>
            <span className="font-mono text-[11px] text-forge-text">
              {status.ahead ?? 0} ahead / {status.behind ?? 0} behind
            </span>
          </div>
          <div className="flex items-center justify-between gap-3">
            <span className="text-[10px] uppercase tracking-[0.12em] text-forge-muted">State</span>
            <span className={`inline-flex items-center gap-1 text-[11px] ${status.is_dirty ? "text-forge-signal" : "text-emerald-400"}`}>
              {status.is_dirty ? <AlertCircle size={12} /> : <CheckCircle2 size={12} />}
              {status.is_dirty ? "Dirty" : "Clean"}
            </span>
          </div>
          {(status.staged ?? []).length > 0 && (
            <div>
              <p className="mb-1 text-[10px] uppercase tracking-[0.12em] text-forge-muted">Staged</p>
              <ul className="space-y-1 text-[11px] text-forge-text">
                {(status.staged ?? []).map((file) => <li key={file} className="truncate font-mono">{file}</li>)}
              </ul>
            </div>
          )}
          {(status.modified ?? []).length > 0 && (
            <div>
              <p className="mb-1 text-[10px] uppercase tracking-[0.12em] text-forge-muted">Modified</p>
              <ul className="space-y-1 text-[11px] text-forge-text">
                {(status.modified ?? []).map((file) => <li key={file} className="truncate font-mono">{file}</li>)}
              </ul>
            </div>
          )}
          {(status.untracked ?? []).length > 0 && (
            <div>
              <p className="mb-1 text-[10px] uppercase tracking-[0.12em] text-forge-muted">Untracked</p>
              <ul className="space-y-1 text-[11px] text-forge-text">
                {(status.untracked ?? []).map((file) => <li key={file} className="truncate font-mono">{file}</li>)}
              </ul>
            </div>
          )}
        </div>
      )}

      {sandboxReady && diffQuery.isError && !statusQuery.isError && (
        <StatusNotice>
          <div className="flex items-start justify-between gap-3">
            <span>{diffQuery.error instanceof Error ? diffQuery.error.message : "Could not load Git diff."}</span>
            <button
              type="button"
              className="shrink-0 underline underline-offset-2 disabled:opacity-50"
              disabled={diffQuery.isFetching}
              onClick={() => void diffQuery.refetch()}
            >
              Retry
            </button>
          </div>
        </StatusNotice>
      )}

      <form onSubmit={onSubmit} className="space-y-3 border border-[var(--border)] bg-forge-bg p-3">
        <label className="block text-[10px] uppercase tracking-[0.12em] text-forge-muted">Commit message</label>
        <textarea
          value={message}
          onChange={(event) => setMessage(event.target.value)}
          rows={3}
          className="w-full resize-none bg-transparent text-xs text-forge-text outline-none placeholder:text-forge-muted"
          placeholder="Describe the workspace changes"
        />
        <p className="text-[10px] leading-4 text-forge-muted">
          Only selected files are included in a commit. Push publishes existing commits without committing other changes.
        </p>
        <div className="flex gap-2">
          <button type="submit" disabled={commitMutation.isPending || revertMutation.isPending || pushMutation.isPending || !status || selectedPaths.length === 0} className="inline-flex flex-1 items-center justify-center gap-2 border border-forge-accent/35 bg-forge-accent/[0.08] px-3 py-2 text-[11px] font-bold uppercase tracking-[0.08em] text-forge-accent disabled:cursor-not-allowed disabled:opacity-50">
            {commitMutation.isPending ? <Loader2 size={12} className="animate-spin" /> : <Send size={12} />}
            Commit
          </button>
          <button type="button" onClick={() => void revertMutation.mutate()} disabled={revertMutation.isPending || commitMutation.isPending || pushMutation.isPending || !status || selectedPaths.length === 0} className="inline-flex flex-1 items-center justify-center gap-2 border border-[var(--border)] px-3 py-2 text-[11px] font-bold uppercase tracking-[0.08em] text-forge-muted disabled:cursor-not-allowed disabled:opacity-50">
            {revertMutation.isPending ? <Loader2 size={12} className="animate-spin" /> : <RotateCcw size={12} />}
            {revertMutation.isPending ? "Reverting" : "Revert selected"}
          </button>
          <button type="button" onClick={() => void pushMutation.mutate()} disabled={pushMutation.isPending || commitMutation.isPending || revertMutation.isPending || !status || (status.ahead ?? 0) === 0} className="inline-flex flex-1 items-center justify-center gap-2 border border-[var(--border)] px-3 py-2 text-[11px] font-bold uppercase tracking-[0.08em] text-forge-muted disabled:cursor-not-allowed disabled:opacity-50">
            {pushMutation.isPending ? <Loader2 size={12} className="animate-spin" /> : <GitBranch size={12} />}
            {pushMutation.isPending ? "Publishing" : "Push to GitHub"}
          </button>
        </div>
      </form>
      </div>
      <section className="flex min-h-0 flex-1 flex-col gap-2">
        <div className="flex shrink-0 items-center justify-between text-[10px] font-semibold uppercase tracking-[0.12em] text-forge-muted">
          <span>
            File changes · {selectionSummary.selectedCount}/{selectionSummary.totalCount} selected
          </span>
          <span className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setFileSelection(resetGitSelection(changedPaths, "all"))}
              disabled={changedPaths.length === 0}
              className="hover:text-forge-text disabled:opacity-40"
            >
              All
            </button>
            <button
              type="button"
              onClick={() => setFileSelection(resetGitSelection(changedPaths, "none"))}
              disabled={changedPaths.length === 0}
              className="hover:text-forge-text disabled:opacity-40"
            >
              None
            </button>
          </span>
        </div>
        {status ? (
          <GitDiffPreview
            accessToken={accessToken}
            sandboxId={sandboxId}
            status={status}
            files={files}
            selectedPaths={selectedPaths}
            onToggleFile={(path, selected) => setFileSelection((current) => nextGitSelection(current, changedPaths, path, selected))}
          />
        ) : (
          <div className="flex min-h-0 flex-1 items-center justify-center border border-[var(--border)] text-xs text-forge-muted">
            {sandboxReady ? "Waiting for Git status" : "Git review is unavailable until the sandbox is running"}
          </div>
        )}
      </section>
      {error && <div className="text-xs text-forge-signal">{error}</div>}
    </div>
  );
}

function StatusNotice({ children }: { children: ReactNode }) {
  return (
    <div className="border border-forge-signal/25 bg-forge-signal/[0.06] p-3 text-[11px] leading-5 text-forge-signal">
      {children}
    </div>
  );
}
