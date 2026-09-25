import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Archive, Boxes, FolderGit2, Plus, RefreshCw, Search } from "lucide-react";
import { useAuth } from "../../../auth/useAuth";
import { useToast } from "../../../../shared/ui/useToast";
import { useWorkspaceId } from "../../hooks/useWorkspaceId";
import { CreateProjectDialog } from "./components/CreateProjectDialog";
import { ImportGitHubRepositoriesDialog } from "./components/ImportGitHubRepositoriesDialog";
import { ProjectCard } from "./components/ProjectCard";
import { syncAllProjects } from "../../api/projects.api";
import { useProjects, useRefreshProjects } from "./hooks/useProjects";
import { PaginationControls } from "../../../../shared/ui/PaginationControls";
import { useDebouncedValue } from "../../../../shared/hooks/useDebouncedValue";

export function ProjectsPage() {
  const workspaceId = useWorkspaceId();
  const { tokens } = useAuth();
  const accessToken = tokens?.access_token ?? "";
  const toast = useToast();
  const [createOpen, setCreateOpen] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebouncedValue(search);
  const projectsQuery = useProjects(accessToken, workspaceId, { page, perPage: 10, search: debouncedSearch });
  const refreshProjects = useRefreshProjects(workspaceId);
  const syncAllMutation = useMutation({
    mutationFn: () => syncAllProjects(accessToken, workspaceId),
    onSuccess: (result) => {
      toast.pushSuccess(`${result.triggered} project${result.triggered === 1 ? "" : "s"} queued for sync${result.failed ? `, ${result.failed} failed` : ""}`);
      void refreshProjects();
    },
    onError: (error) => toast.pushError(error instanceof Error ? error.message : "Could not sync workspace projects"),
  });
  const projects = projectsQuery.data?.items ?? [];

  return (
    <div className="animate-page-enter space-y-6">
      <header className="flex flex-col justify-between gap-5 border-b border-white/[0.08] pb-6 sm:flex-row sm:items-end">
        <div>
          <span className="mb-3 block font-mono text-[10px] uppercase tracking-[.14em] text-forge-accent">Project workspace</span>
          <h1 className="m-0 text-3xl font-extrabold tracking-[-.05em] text-forge-text sm:text-[42px]">Projects</h1>
          <p className="mt-3 max-w-xl text-sm leading-6 text-forge-muted">Organize repositories and agent work into focused project spaces.</p>
        </div>
        <div className="flex gap-2">
          <button className="grid size-11 place-items-center border border-white/[0.1] text-forge-muted transition hover:border-forge-accent/40 hover:text-forge-text disabled:opacity-50" onClick={() => void refreshProjects()} disabled={projectsQuery.isFetching} aria-label="Refresh projects" title="Refresh projects"><RefreshCw size={16} className={projectsQuery.isFetching ? "animate-spin" : ""} /></button>
          <button className="inline-flex items-center gap-2 border border-forge-accent/35 px-4 py-3 text-xs font-extrabold text-forge-accent transition hover:bg-forge-accent/[0.08] disabled:opacity-50" onClick={() => syncAllMutation.mutate()} disabled={syncAllMutation.isPending}><RefreshCw size={16} className={syncAllMutation.isPending ? "animate-spin" : ""} /> Sync all</button>
          <button className="inline-flex items-center gap-2 border border-forge-accent/35 px-4 py-3 text-xs font-extrabold text-forge-accent transition hover:bg-forge-accent/[0.08]" onClick={() => setImportOpen(true)}><FolderGit2 size={16} /> Import GitHub</button>
          <button className="inline-flex items-center gap-2 bg-forge-accent px-4 py-3 text-xs font-extrabold text-forge-bg transition hover:bg-forge-accent-strong" onClick={() => setCreateOpen(true)}><Plus size={16} /> New project</button>
        </div>
      </header>

      <section className="grid gap-3 sm:grid-cols-3">
        <Summary icon={Boxes} label="Total projects" value={projects.length} />
        <Summary icon={FolderGit2} label="Connected repositories" value={projects.filter((project) => project.repository).length} />
        <Summary icon={Archive} label="Active projects" value={projects.filter((project) => project.status === "ACTIVE").length} />
      </section>

      <label className="relative block max-w-[560px]">
        <Search size={16} className={`pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-forge-muted ${search !== debouncedSearch ? "animate-pulse text-forge-accent" : ""}`} />
        <input className="input pl-10" value={search} onChange={(event) => { setSearch(event.target.value); setPage(1); }} placeholder="Search project or repository" aria-label="Search projects and repositories" />
      </label>
      {(search !== debouncedSearch || projectsQuery.isPlaceholderData) && <div className="flex items-center gap-2 text-xs text-forge-muted" role="status"><span className="size-1.5 animate-ping rounded-full bg-forge-accent" />Filtering projects...</div>}

      {projectsQuery.isLoading ? <Loading /> : projectsQuery.isError ? (
        <div className="border border-forge-signal/30 bg-forge-signal/[0.06] p-5 text-sm text-forge-soft">Could not load projects. {projectsQuery.error instanceof Error ? projectsQuery.error.message : "Try again."}</div>
      ) : projects.length === 0 ? (
        <EmptyProjects onCreate={() => setCreateOpen(true)} />
      ) : (
        <section className="grid gap-3" aria-busy={projectsQuery.isFetching}>{projects.map((project) => <ProjectCard key={project.id} project={project} accessToken={accessToken} workspaceId={workspaceId} onError={(message) => toast.pushError(message)} />)}<PaginationControls page={projectsQuery.data?.page ?? page} totalPages={projectsQuery.data?.total_pages ?? 0} total={projectsQuery.data?.total ?? 0} onPageChange={setPage} /></section>
      )}

      {createOpen && <CreateProjectDialog accessToken={accessToken} workspaceId={workspaceId} onClose={() => setCreateOpen(false)} />}
      {importOpen && <ImportGitHubRepositoriesDialog accessToken={accessToken} workspaceId={workspaceId} onClose={() => setImportOpen(false)} />}
    </div>
  );
}

function Summary({ icon: Icon, label, value }: { icon: typeof Boxes; label: string; value: number }) {
  return <div className="border border-white/[0.09] bg-forge-panel/70 p-5"><Icon size={17} className="text-forge-accent" /><p className="mt-4 text-2xl font-extrabold text-forge-text">{value}</p><p className="mt-1 text-xs text-forge-muted">{label}</p></div>;
}

function EmptyProjects({ onCreate }: { onCreate: () => void }) {
  return <div className="border border-dashed border-white/[0.14] bg-forge-panel/60 px-6 py-16 text-center"><FolderGit2 className="mx-auto text-forge-accent" size={28} /><h2 className="mt-4 text-xl font-bold text-forge-text">No projects yet</h2><p className="mx-auto mt-2 max-w-md text-sm leading-6 text-forge-muted">Create an empty project or connect an existing GitHub repository to get started.</p><button className="mt-6 bg-forge-accent px-4 py-3 text-xs font-extrabold text-forge-bg" onClick={onCreate}>Create your first project</button></div>;
}

function Loading() {
  return <div className="grid min-h-[280px] place-content-center gap-3 text-sm text-forge-muted"><span className="mx-auto size-2.5 animate-pulse bg-forge-accent" />Loading projects</div>;
}
