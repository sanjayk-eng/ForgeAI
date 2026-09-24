import { useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, X } from "lucide-react";
import { useToast } from "../../../../../shared/ui/useToast";
import { resolveRepository, syncProject, updateProject, updateRepositoryBranch } from "../../../api/projects.api";
import type { Project } from "../types/project.types";
import { LoadingOverlay } from "../../../../../shared/ui/LoadingOverlay";

type EditProjectDialogProps = {
  accessToken: string;
  workspaceId: string;
  project: Project;
  onClose: () => void;
};

export function EditProjectDialog({ accessToken, workspaceId, project, onClose }: EditProjectDialogProps) {
  const [name, setName] = useState(project.name);
  const [description, setDescription] = useState(project.description ?? "");
  const [branch, setBranch] = useState(project.repository?.default_branch ?? "");
  const toast = useToast();
  const queryClient = useQueryClient();
  const branchesQuery = useQuery({
    queryKey: ["project-branches", project.id, project.repository?.repository_url],
    queryFn: () => resolveRepository(accessToken, project.repository!.repository_url),
    enabled: Boolean(project.repository),
  });
  const mutation = useMutation({
    mutationFn: async () => {
      await updateProject(accessToken, project.id, {
        name: name.trim(),
        description: description.trim() || undefined,
        status: project.status,
      });
      if (project.repository && branch.trim() && branch.trim() !== project.repository.default_branch) {
        await updateRepositoryBranch(accessToken, project.id, branch.trim());
        await syncProject(accessToken, project.id);
      }
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects", workspaceId] });
      toast.pushSuccess("Project updated");
      onClose();
    },
    onError: (error) => toast.pushError(error instanceof Error ? error.message : "Could not update project"),
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    if (name.trim()) mutation.mutate();
  }

  return (
    <div className="fixed inset-0 z-30 grid place-items-center bg-[var(--overlay)] p-5 backdrop-blur-md">
      <form className="relative w-full max-w-[520px] rounded-xl border border-forge-accent/20 bg-forge-card p-7 shadow-2xl" onSubmit={submit}>
        {mutation.isPending && <LoadingOverlay message="Updating project" />}
        <div className="flex items-start justify-between gap-5"><div><span className="mb-2 block font-mono text-[10px] uppercase tracking-[.12em] text-forge-accent">Project settings</span><h2 className="m-0 text-[22px] font-bold tracking-[-.03em]">Edit project</h2></div><button type="button" className="grid size-9 place-items-center rounded-md text-forge-muted hover:bg-[var(--surface-hover)]" onClick={onClose} aria-label="Close dialog"><X size={18} /></button></div>
        <div className="mt-6 grid gap-4"><label className="block"><span className="label">Project name</span><input className="input mt-2" value={name} onChange={(event) => setName(event.target.value)} autoFocus /></label><label className="block"><span className="label">Description</span><textarea className="input mt-2 min-h-24 resize-y" value={description} onChange={(event) => setDescription(event.target.value)} /></label>{project.repository && <label className="block"><span className="label">GitHub branch</span>{branchesQuery.isLoading ? <div className="mt-2 flex items-center gap-2 text-xs text-forge-muted"><LoaderCircle size={15} className="animate-spin text-forge-accent" />Loading branches</div> : <><select className="input mt-2" value={branch} onChange={(event) => setBranch(event.target.value)} disabled={branchesQuery.isError}><option value={project.repository.default_branch}>{project.repository.default_branch}</option>{(branchesQuery.data?.branches ?? []).filter((item) => item !== project.repository?.default_branch).map((item) => <option key={item} value={item}>{item}</option>)}</select>{branchesQuery.isError && <span className="mt-1 block text-xs text-forge-muted">Branches could not be loaded; the current branch remains selected.</span>}</>}</label>}</div>
        <div className="mt-7 flex justify-end gap-2"><button type="button" className="border border-white/[0.1] px-4 py-3 text-xs font-bold text-forge-muted hover:text-forge-text" onClick={onClose}>Cancel</button><button type="submit" className="inline-flex items-center gap-2 bg-forge-accent px-4 py-3 text-xs font-extrabold text-forge-bg disabled:opacity-50" disabled={!name.trim() || mutation.isPending}>{mutation.isPending && <LoaderCircle size={15} className="animate-spin" />}Save changes</button></div>
      </form>
    </div>
  );
}
