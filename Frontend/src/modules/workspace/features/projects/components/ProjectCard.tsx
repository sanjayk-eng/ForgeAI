import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CheckCircle2, Edit3, FolderGit2, LoaderCircle, RefreshCw, Trash2 } from "lucide-react";
import { useState } from "react";
import { deleteProject, resolveRepository, syncProject, updateRepositoryBranch } from "../../../api/projects.api";
import type { Project } from "../types/project.types";
import { EditProjectDialog } from "./EditProjectDialog";

const syncStyles = {
  PENDING: "text-amber-200",
  SYNCING: "text-sky-200",
  SYNCED: "text-forge-accent",
  FAILED: "text-forge-signal",
} as const;

export function ProjectCard({
  project,
  accessToken,
  workspaceId,
  onError,
  onDeleted,
}: {
  project: Project;
  accessToken: string;
  workspaceId: string;
  onError: (message: string) => void;
  onDeleted: () => void;
}) {
  const repository = project.repository;
  const [editOpen, setEditOpen] = useState(false);
  const queryClient = useQueryClient();
  const syncMutation = useMutation({
    mutationFn: () => syncProject(accessToken, project.id),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["projects", workspaceId] }),
    onError: (error) => onError(error instanceof Error ? error.message : "Could not start repository sync"),
  });
  const syncing = repository?.sync_status === "SYNCING" || syncMutation.isPending;
  const deleteMutation = useMutation({
    mutationFn: () => deleteProject(accessToken, project.id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects", workspaceId] });
      onDeleted();
    },
    onError: (error) => onError(error instanceof Error ? error.message : "Could not remove project"),
  });

  function remove() {
    if (deleteMutation.isPending || !window.confirm(`Remove project "${project.name}"? This cannot be undone.`)) return;
    deleteMutation.mutate();
  }

  return (
    <article className={`relative overflow-hidden border bg-forge-panel/70 p-5 transition sm:p-6 ${syncing ? "border-sky-300/35" : "border-[var(--border)] hover:border-forge-accent/25"}`}>
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
        <div className="flex min-w-0 items-start gap-3"><span className={`grid size-10 shrink-0 place-items-center border bg-forge-accent/[0.07] text-forge-accent ${syncing ? "animate-pulse border-sky-300/40 text-sky-200" : "border-forge-accent/20"}`}><FolderGit2 size={19} /></span><div className="min-w-0"><h2 className="truncate text-base font-bold text-forge-text">{project.name}</h2><p className="mt-1 font-mono text-[11px] text-forge-muted">/{project.slug}</p></div></div>
        <div className="flex items-center gap-2 font-mono text-[10px] uppercase tracking-[.08em]"><span className="border border-[var(--border)] px-2 py-1 text-forge-muted">{project.type}</span>{repository && <span className={`inline-flex items-center gap-1 ${syncStyles[repository.sync_status]}`}>{syncing && <RefreshCw size={11} className="animate-spin" />}{repository.sync_status === "SYNCED" && <CheckCircle2 size={11} />}{repository.sync_status}</span>}<button type="button" className="grid size-7 place-items-center border border-[var(--border)] text-forge-muted transition hover:border-forge-accent/40 hover:text-forge-text" onClick={() => setEditOpen(true)} aria-label={`Edit ${project.name}`} title="Edit project"><Edit3 size={13} /></button><button type="button" className="grid size-7 place-items-center border border-[var(--border)] text-forge-muted transition hover:border-forge-signal/50 hover:text-forge-signal disabled:cursor-not-allowed disabled:opacity-50" onClick={remove} disabled={deleteMutation.isPending} aria-label={`Remove ${project.name}`} title="Remove project"><Trash2 size={13} /></button></div>
      </div>
      {project.description && <p className="mt-5 max-w-2xl text-sm leading-6 text-forge-muted">{project.description}</p>}
      {repository ? <RepositoryFooter accessToken={accessToken} repository={repository} syncing={syncing} onSync={() => syncMutation.mutate()} onError={onError} /> : <div className="mt-5 border-t border-[var(--border)] pt-4 text-xs text-forge-muted">Empty project · repository can be connected later</div>}
      {syncing && <div className="absolute inset-x-0 bottom-0 h-0.5 overflow-hidden bg-sky-300/10"><div className="h-full w-1/3 animate-[sync-progress_1.4s_ease-in-out_infinite] bg-sky-300" /></div>}
      {editOpen && <EditProjectDialog accessToken={accessToken} workspaceId={workspaceId} project={project} onClose={() => setEditOpen(false)} />}
    </article>
  );
}

function RepositoryFooter({ accessToken, repository, syncing, onSync, onError }: { accessToken: string; repository: NonNullable<Project["repository"]>; syncing: boolean; onSync: () => void; onError: (message: string) => void }) {
  return <div className="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-[var(--border)] pt-4 text-xs text-forge-soft"><div className="flex flex-wrap items-center gap-x-5 gap-y-2"><span className="inline-flex items-center gap-2"><FolderGit2 size={14} className="text-forge-accent" />{repository.github_owner}/{repository.github_repository_name}</span><BranchEditor accessToken={accessToken} repository={repository} syncing={syncing} onError={onError} /></div><button className="inline-flex items-center gap-2 border border-[var(--border)] px-3 py-2 text-[10px] font-extrabold uppercase tracking-[.08em] text-forge-soft transition hover:border-forge-accent/40 hover:text-forge-text disabled:cursor-not-allowed disabled:opacity-50" onClick={onSync} disabled={syncing} title="Sync repository"><RefreshCw size={13} className={syncing ? "animate-spin" : ""} />{syncing ? "Syncing" : repository.sync_status === "FAILED" ? "Retry sync" : "Sync now"}</button></div>;
}

function BranchEditor({ accessToken, repository, syncing, onError }: { accessToken: string; repository: NonNullable<Project["repository"]>; syncing: boolean; onError: (message: string) => void }) {
  const [selectedBranch, setSelectedBranch] = useState(repository.default_branch);
  const queryClient = useQueryClient();
  const branchesQuery = useQuery({
    queryKey: ["project-branches", repository.project_id, repository.repository_url],
    queryFn: () => resolveRepository(accessToken, repository.repository_url),
  });
  const branchMutation = useMutation({
    mutationFn: (branch: string) => updateRepositoryBranch(accessToken, repository.project_id, branch),
    onSuccess: (_, branch) => {
      setSelectedBranch(branch);
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
      void syncProject(accessToken, repository.project_id).catch((error) => {
        onError(error instanceof Error ? error.message : "Branch updated, but repository sync could not start");
      });
    },
    onError: (error) => onError(error instanceof Error ? error.message : "Could not update repository branch"),
  });

  if (branchesQuery.isLoading) return <span className="inline-flex items-center gap-2 font-mono text-forge-muted"><LoaderCircle size={13} className="animate-spin" />Loading branch</span>;
  if (branchesQuery.isError) return <span className="font-mono text-forge-muted">{repository.default_branch}</span>;
  return <label className="inline-flex items-center gap-2"><span className="sr-only">GitHub branch</span><select className="input min-w-36 py-1 font-mono text-[11px]" value={selectedBranch} onChange={(event) => branchMutation.mutate(event.target.value)} disabled={syncing || branchMutation.isPending}><option value={selectedBranch}>{selectedBranch}</option>{(branchesQuery.data?.branches ?? []).filter((branch) => branch !== selectedBranch).map((branch) => <option key={branch} value={branch}>{branch}</option>)}</select></label>;
}
