import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Check, GitBranch, LoaderCircle, X } from "lucide-react";
import { useToast } from "../../../../../shared/ui/useToast";
import {
  importGitHubRepositories,
  listGitHubRepositories,
} from "../../../api/projects.api";
import type { GitHubRepositoryOption } from "../types/project.types";
import { LoadingOverlay } from "../../../../../shared/ui/LoadingOverlay";

type ImportGitHubRepositoriesDialogProps = {
  accessToken: string;
  workspaceId: string;
  onClose: () => void;
};

export function ImportGitHubRepositoriesDialog({
  accessToken,
  workspaceId,
  onClose,
}: ImportGitHubRepositoriesDialogProps) {
  const [organization, setOrganization] = useState("all");
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const toast = useToast();
  const queryClient = useQueryClient();
  const repositoriesQuery = useQuery({
    queryKey: ["github-repositories", workspaceId],
    queryFn: () => listGitHubRepositories(accessToken, workspaceId),
    enabled: Boolean(accessToken && workspaceId),
  });
  const importMutation = useMutation({
    mutationFn: () =>
      importGitHubRepositories(
        accessToken,
        workspaceId,
        (catalog?.repositories ?? []).filter((repository) => selected.has(repository.github_repository_id)),
      ),
    onSuccess: (result) => {
      void queryClient.invalidateQueries({ queryKey: ["projects", workspaceId] });
      toast.pushSuccess(`${result.projects.length} project${result.projects.length === 1 ? "" : "s"} imported`);
      onClose();
    },
    onError: (error) =>
      toast.pushError(error instanceof Error ? error.message : "Could not import repositories"),
  });

  const catalog = repositoriesQuery.data;
  const visibleRepositories = useMemo(
    () =>
      catalog?.repositories.filter(
        (repository) => organization === "all" || repository.organization === organization,
      ) ?? [],
    [catalog, organization],
  );
  const allVisibleSelected =
    visibleRepositories.length > 0 &&
    visibleRepositories.every((repository) => selected.has(repository.github_repository_id));

  function toggle(repository: GitHubRepositoryOption) {
    setSelected((current) => {
      const next = new Set(current);
      if (next.has(repository.github_repository_id)) next.delete(repository.github_repository_id);
      else next.add(repository.github_repository_id);
      return next;
    });
  }

  function toggleVisible() {
    setSelected((current) => {
      const next = new Set(current);
      if (allVisibleSelected) visibleRepositories.forEach((repository) => next.delete(repository.github_repository_id));
      else visibleRepositories.forEach((repository) => next.add(repository.github_repository_id));
      return next;
    });
  }

  return (
    <div className="fixed inset-0 z-30 grid place-items-center bg-[var(--overlay)] p-5 backdrop-blur-md">
      <section className="relative flex max-h-[min(760px,calc(100vh-40px))] w-full max-w-[720px] flex-col rounded-xl border border-forge-accent/20 bg-forge-card shadow-2xl">
        {(repositoriesQuery.isLoading || importMutation.isPending) && <LoadingOverlay message={importMutation.isPending ? "Importing repositories" : "Loading GitHub repositories"} />}
        <header className="flex items-start justify-between gap-5 border-b border-white/[0.08] p-7">
          <div>
            <span className="mb-2 block font-mono text-[10px] uppercase tracking-[.12em] text-forge-accent">GitHub import</span>
            <h2 className="m-0 text-[22px] font-bold tracking-[-.03em]">Load repositories</h2>
            <p className="mt-2 text-sm text-forge-muted">Choose repositories to create as projects and sync.</p>
          </div>
          <button type="button" className="grid size-9 place-items-center rounded-md text-forge-muted hover:bg-[var(--surface-hover)]" onClick={onClose} aria-label="Close dialog"><X size={18} /></button>
        </header>

        {repositoriesQuery.isLoading ? (
          <div className="grid min-h-[280px] place-content-center gap-3 text-sm text-forge-muted"><LoaderCircle className="mx-auto animate-spin text-forge-accent" size={20} />Loading GitHub repositories</div>
        ) : repositoriesQuery.isError ? (
          <div className="m-7 border border-forge-signal/30 bg-forge-signal/[0.06] p-5 text-sm text-forge-soft">Could not load GitHub repositories. {repositoriesQuery.error instanceof Error ? repositoriesQuery.error.message : "Connect GitHub and try again."}</div>
        ) : (
          <>
            <div className="flex flex-wrap items-center gap-3 border-b border-white/[0.08] p-5">
              <label className="min-w-52 flex-1"><span className="label">Source</span><select className="input mt-2" value={organization} onChange={(event) => setOrganization(event.target.value)}><option value="all">All organizations and personal</option><option value="personal">Personal repositories</option>{catalog?.organizations.map((item) => <option key={item} value={item}>{item}</option>)}</select></label>
              <button type="button" className="mt-5 inline-flex items-center gap-2 border border-white/[0.12] px-3 py-3 text-xs font-bold text-forge-soft hover:border-forge-accent/40" onClick={toggleVisible} disabled={visibleRepositories.length === 0}>{allVisibleSelected ? "Clear visible" : "Select visible"}</button>
            </div>
            <div className="min-h-0 flex-1 overflow-y-auto p-5">
              {visibleRepositories.length === 0 ? <p className="py-12 text-center text-sm text-forge-muted">No repositories found for this source.</p> : <div className="grid gap-2">{visibleRepositories.map((repository) => { const checked = selected.has(repository.github_repository_id); return <button key={repository.github_repository_id} type="button" className={`flex items-center gap-3 border p-3 text-left transition ${checked ? "border-forge-accent/45 bg-forge-accent/[0.08]" : "border-white/[0.09] hover:border-white/20"}`} onClick={() => toggle(repository)}><span className={`grid size-7 shrink-0 place-items-center border ${checked ? "border-forge-accent bg-forge-accent text-forge-bg" : "border-white/[0.16] text-transparent"}`}><Check size={15} /></span><GitBranch size={16} className="shrink-0 text-forge-muted" /><span className="min-w-0 flex-1"><strong className="block truncate text-sm text-forge-soft">{repository.github_owner}/{repository.github_repository_name}</strong><span className="mt-1 block text-xs text-forge-muted">{repository.organization === "personal" ? "Personal" : repository.organization} · {repository.default_branch}</span></span></button>; })}</div>}
            </div>
          </>
        )}

        <footer className="flex items-center justify-between gap-4 border-t border-white/[0.08] p-5"><span className="text-xs text-forge-muted">{selected.size} selected</span><div className="flex gap-2"><button type="button" className="border border-white/[0.1] px-4 py-3 text-xs font-bold text-forge-muted hover:text-forge-text" onClick={onClose}>Cancel</button><button type="button" className="inline-flex items-center gap-2 bg-forge-accent px-4 py-3 text-xs font-extrabold text-forge-bg disabled:opacity-50" onClick={() => importMutation.mutate()} disabled={selected.size === 0 || importMutation.isPending}>{importMutation.isPending && <LoaderCircle size={15} className="animate-spin" />}Import selected</button></div></footer>
      </section>
    </div>
  );
}
