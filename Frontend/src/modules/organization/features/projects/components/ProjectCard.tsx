import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Bot, Edit3, FolderGit2, LoaderCircle, Trash2 } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { deleteProject, resolveRepository, updateRepositoryBranch } from "../../../api/projects.api";
import { useSandbox } from "../hooks/useSandbox";
import type { Project } from "../types/project.types";
import { EditProjectDialog } from "./EditProjectDialog";
import { SandboxStatus } from "./SandboxStatus";

export function ProjectCard({
  project,
  accessToken,
  organizationId,
  onError,
  onDeleted,
}: {
  project: Project;
  accessToken: string;
  organizationId: string;
  onError: (message: string) => void;
  onDeleted: () => void;
}) {
  const navigate = useNavigate();
  const repository = project.repository;
  const [editOpen, setEditOpen] = useState(false);
  const queryClient = useQueryClient();
  const { sandbox, isLoading: sandboxLoading } = useSandbox(accessToken, project.id);
  
  const deleteMutation = useMutation({
    mutationFn: () => deleteProject(accessToken, project.id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects", organizationId] });
      onDeleted();
    },
    onError: (error: unknown) => onError(error instanceof Error ? error.message : "Could not remove project"),
  });

  function remove() {
    if (deleteMutation.isPending || !window.confirm(`Remove project "${project.name}"? This cannot be undone.`)) return;
    deleteMutation.mutate();
  }

  const handleOpenProject = () => {
    navigate(`/organizations/${organizationId}/projects/${project.id}`);
  };
  const handleOpenAgent = () => {
    navigate(`/organizations/${organizationId}/projects/${project.id}?panel=agent`);
  };

  return (
    <article className="relative overflow-hidden border border-[var(--border)] bg-forge-panel/70 p-5 transition hover:border-forge-accent/25 sm:p-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
        <button
          onClick={handleOpenProject}
          className="flex min-w-0 items-start gap-3 text-left hover:opacity-80 transition"
        >
          <span className="grid size-10 shrink-0 place-items-center border border-forge-accent/20 bg-forge-accent/[0.07] text-forge-accent">
            <FolderGit2 size={19} />
          </span>
          <div className="min-w-0">
            <h2 className="truncate text-base font-bold text-forge-text">{project.name}</h2>
            <p className="mt-1 font-mono text-[11px] text-forge-muted">/{project.slug}</p>
          </div>
        </button>
        <div className="flex items-center gap-2 font-mono text-[10px] uppercase tracking-[.08em]">
          <button type="button" className="inline-flex h-7 items-center gap-1.5 border border-forge-accent/25 px-2 text-forge-accent transition hover:bg-forge-accent/[0.08]" onClick={handleOpenAgent} aria-label={`Open agent for ${project.name}`} title="Open project agent"><Bot size={13} /> Agent</button>
          {sandbox && <SandboxStatus status={sandbox.status} showLabel={false} />}
          {!sandbox && !sandboxLoading && <span className="text-xs text-forge-muted">Setting up...</span>}
          <span className="border border-[var(--border)] px-2 py-1 text-forge-muted">{project.type}</span>
          <button type="button" className="grid size-7 place-items-center border border-[var(--border)] text-forge-muted transition hover:border-forge-accent/40 hover:text-forge-text" onClick={() => setEditOpen(true)} aria-label={`Edit ${project.name}`} title="Edit project"><Edit3 size={13} /></button>
          <button type="button" className="grid size-7 place-items-center border border-[var(--border)] text-forge-muted transition hover:border-forge-signal/50 hover:text-forge-signal disabled:cursor-not-allowed disabled:opacity-50" onClick={remove} disabled={deleteMutation.isPending} aria-label={`Remove ${project.name}`} title="Remove project"><Trash2 size={13} /></button>
        </div>
      </div>
      {project.description && <p className="mt-5 max-w-2xl text-sm leading-6 text-forge-muted">{project.description}</p>}
      {repository ? <RepositoryFooter accessToken={accessToken} repository={repository} onError={onError} /> : <div className="mt-5 border-t border-[var(--border)] pt-4 text-xs text-forge-muted">Empty project · repository can be connected later</div>}
      {editOpen && <EditProjectDialog accessToken={accessToken} organizationId={organizationId} project={project} onClose={() => setEditOpen(false)} />}
    </article>
  );
}

function RepositoryFooter({ accessToken, repository, onError }: { accessToken: string; repository: NonNullable<Project["repository"]>; onError: (message: string) => void }) {
  return <div className="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-[var(--border)] pt-4 text-xs text-forge-soft"><div className="flex flex-wrap items-center gap-x-5 gap-y-2"><span className="inline-flex items-center gap-2"><FolderGit2 size={14} className="text-forge-accent" />{repository.github_owner}/{repository.github_repository_name}</span><BranchEditor accessToken={accessToken} repository={repository} onError={onError} /></div></div>;
}

function BranchEditor({ accessToken, repository, onError }: { accessToken: string; repository: NonNullable<Project["repository"]>; onError: (message: string) => void }) {
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
    },
    onError: (error: unknown) => onError(error instanceof Error ? error.message : "Could not update repository branch"),
  });

  if (branchesQuery.isLoading) return <span className="inline-flex items-center gap-2 font-mono text-forge-muted"><LoaderCircle size={13} className="animate-spin" />Loading branch</span>;
  if (branchesQuery.isError) return <span className="font-mono text-forge-muted">{repository.default_branch}</span>;
  return <label className="inline-flex items-center gap-2"><span className="sr-only">GitHub branch</span><select className="input min-w-36 py-1 font-mono text-[11px]" value={selectedBranch} onChange={(event) => branchMutation.mutate(event.target.value)} disabled={branchMutation.isPending}><option value={selectedBranch}>{selectedBranch}</option>{(branchesQuery.data?.branches ?? []).filter((branch) => branch !== selectedBranch).map((branch) => <option key={branch} value={branch}>{branch}</option>)}</select></label>;
}
