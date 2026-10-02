import { useState, type FormEvent, type ReactNode } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check, GitBranch, Globe2, Link2, LoaderCircle, LockKeyhole, X } from "lucide-react";
import { useToast } from "../../../../../shared/ui/useToast";
import { createProject, resolveRepository } from "../../../api/projects.api";
import type { CreateProjectInput, ProjectType, ResolvedRepository } from "../types/project.types";

export function CreateProjectDialog({ accessToken, workspaceId, onClose }: { accessToken: string; workspaceId: string; onClose: () => void }) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [type, setType] = useState<ProjectType>("REPOSITORY");
  const [repositoryUrl, setRepositoryUrl] = useState("");
  const [resolvedRepository, setResolvedRepository] = useState<ResolvedRepository>();
  const [selectedBranch, setSelectedBranch] = useState("");
  const toast = useToast();
  const queryClient = useQueryClient();
  const resolveMutation = useMutation({
    mutationFn: () => resolveRepository(accessToken, repositoryUrl.trim()),
    onSuccess: (repository) => {
      setResolvedRepository(repository);
      setSelectedBranch(repository.repository.default_branch);
      setName((current) => current.trim() || repository.repository.github_repository_name);
      toast.pushSuccess("GitHub repository verified");
    },
    onError: (error) => {
      setResolvedRepository(undefined);
      toast.pushError(error instanceof Error ? error.message : "Could not verify repository URL");
    },
  });
  const createMutation = useMutation({
    mutationFn: () => {
      const input: CreateProjectInput = {
        name: name.trim(),
        description: description.trim() || undefined,
        type,
        repository: resolvedRepository ? { ...resolvedRepository.repository, default_branch: selectedBranch } : undefined,
      };
      return createProject(accessToken, workspaceId, input);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects", workspaceId] });
      toast.pushSuccess("Project created");
      onClose();
    },
    onError: (error) => toast.pushError(error instanceof Error ? error.message : "Could not create project"),
  });

  function changeType(nextType: ProjectType) {
    setType(nextType);
    if (nextType === "EMPTY") {
      setResolvedRepository(undefined);
      setSelectedBranch("");
    }
  }

  function submit(event: FormEvent) {
    event.preventDefault();
    if (name.trim() && (type === "EMPTY" || resolvedRepository)) createMutation.mutate();
  }

  const canCreate = Boolean(name.trim() && (type === "EMPTY" || resolvedRepository));

  return (
    <div className="fixed inset-0 z-30 grid place-items-center bg-[var(--overlay)] p-5 backdrop-blur-md">
      <form className="w-full max-w-[560px] rounded-xl border border-[var(--border)] bg-forge-panel p-7 shadow-[0_24px_80px_rgba(20,35,25,.2)]" onSubmit={submit}>
        <div className="flex items-start justify-between gap-5">
          <div><span className="mb-2 block font-mono text-[10px] uppercase tracking-[.12em] text-forge-accent">New project</span><h2 className="m-0 text-[22px] font-bold tracking-[-.03em]">Create a project</h2></div>
          <button type="button" className="grid size-9 place-items-center rounded-md text-forge-muted hover:bg-[var(--surface-hover)]" onClick={onClose} aria-label="Close dialog"><X size={18} /></button>
        </div>
        <div className="mt-6 grid gap-4">
          <Field label="Project name"><input className="input" value={name} onChange={(event) => setName(event.target.value)} placeholder="e.g. Agent Console" autoFocus /></Field>
          <Field label="Description"><textarea className="input min-h-20 resize-y" value={description} onChange={(event) => setDescription(event.target.value)} placeholder="What is this project for?" /></Field>
          <div><span className="label">Project type</span><div className="mt-2 grid grid-cols-2 gap-2">{(["REPOSITORY", "EMPTY"] as const).map((option) => <button key={option} type="button" className={`border px-3 py-3 text-left text-xs font-bold transition ${type === option ? "border-forge-accent bg-forge-accent/[0.1] text-forge-text" : "border-[var(--border)] text-forge-muted hover:border-forge-accent/40 hover:bg-[var(--surface-hover)]"}`} onClick={() => changeType(option)}>{option === "REPOSITORY" ? "GitHub repository" : "Empty project"}<span className="mt-1 block text-[10px] font-normal text-forge-muted">{option === "REPOSITORY" ? "Resolve existing code from GitHub" : "Connect a repository later"}</span></button>)}</div></div>
          {type === "REPOSITORY" && <RepositoryResolver url={repositoryUrl} onUrlChange={(value) => { setRepositoryUrl(value); setResolvedRepository(undefined); setSelectedBranch(""); }} onResolve={() => resolveMutation.mutate()} isResolving={resolveMutation.isPending} repository={resolvedRepository} branch={selectedBranch} onBranchChange={setSelectedBranch} />}
        </div>
        <div className="mt-7 flex justify-end gap-2"><button type="button" className="rounded-md bg-[var(--surface-subtle)] px-4 py-2.5 text-xs font-extrabold text-forge-soft hover:bg-[var(--surface-hover)]" onClick={onClose}>Cancel</button><button className="rounded-md bg-forge-accent px-4 py-2.5 text-xs font-extrabold text-[var(--primary-foreground)] hover:bg-forge-accent-strong disabled:cursor-not-allowed disabled:opacity-45" disabled={!canCreate || createMutation.isPending}>{createMutation.isPending ? "Creating..." : "Create project"}</button></div>
      </form>
    </div>
  );
}

function RepositoryResolver({ url, onUrlChange, onResolve, isResolving, repository, branch, onBranchChange }: { url: string; onUrlChange: (value: string) => void; onResolve: () => void; isResolving: boolean; repository?: ResolvedRepository; branch: string; onBranchChange: (value: string) => void }) {
  const metadata = repository?.repository;
  return <div className="grid gap-4 border-l-2 border-forge-accent/40 pl-4"><div className="flex items-center gap-2 text-xs font-bold text-forge-soft"><GitBranch size={16} /> GitHub repository</div><div><Field label="Repository URL"><div className="flex gap-2"><input className="input mt-0 min-w-0" type="url" value={url} onChange={(event) => onUrlChange(event.target.value)} placeholder="https://github.com/org/repository" /><button type="button" className="grid size-12 shrink-0 place-items-center border border-forge-accent/30 text-forge-accent transition hover:bg-forge-accent/[0.1] disabled:opacity-50" onClick={onResolve} disabled={isResolving || !url.trim()} aria-label="Resolve GitHub repository" title="Resolve GitHub repository">{isResolving ? <LoaderCircle size={17} className="animate-spin" /> : <Link2 size={17} />}</button></div></Field></div>{metadata && <div className="grid gap-3 border border-forge-accent/20 bg-forge-accent/[0.05] p-4"><div className="flex items-center gap-2 text-xs font-bold text-forge-accent"><Check size={15} /> Repository verified</div><div className="grid gap-3 text-xs sm:grid-cols-2"><Meta label="Repository" value={`${metadata.github_owner}/${metadata.github_repository_name}`} /><Meta label="Visibility" value={metadata.private ? "Private" : "Public"} icon={metadata.private ? <LockKeyhole size={12} /> : <Globe2 size={12} />} /><Meta label="GitHub ID" value={String(metadata.github_repository_id)} /><Meta label="Canonical URL" value={metadata.repository_url} /><label><span className="block font-mono text-[10px] text-forge-muted">Sync branch</span><select className="input mt-1" value={branch} onChange={(event) => onBranchChange(event.target.value)}><option value={metadata.default_branch}>{metadata.default_branch}</option>{repository.branches.filter((item) => item !== metadata.default_branch).map((item) => <option key={item} value={item}>{item}</option>)}</select></label></div></div>}</div>;
}

function Meta({ label, value, icon }: { label: string; value: string; icon?: ReactNode }) { return <div><span className="block font-mono text-[10px] text-forge-muted">{label}</span><span className="mt-1 flex items-center gap-1 truncate text-forge-soft" title={value}>{icon}{value}</span></div>; }
function Field({ label, children }: { label: string; children: ReactNode }) { return <label className="block"><span className="label">{label}</span>{children}</label>; }
