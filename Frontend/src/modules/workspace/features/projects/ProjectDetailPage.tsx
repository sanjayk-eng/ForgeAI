import { ArrowLeft, Bot, Code2, FileText, FolderGit2, Loader2 } from "lucide-react";
import { lazy, Suspense, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { useAuth } from "../../../auth/useAuth";
import { useWorkspaceId } from "../../hooks/useWorkspaceId";
import { useSandbox } from "./hooks/useSandbox";
import { SandboxStatus } from "./components/SandboxStatus";
import { FileExplorer } from "./components/files/FileExplorer";
import { AgentPanel } from "./components/AgentPanel";
import { GitPanel } from "./components/GitPanel";
import { ProjectSidePanel } from "./components/ProjectSidePanel";

const CodeEditor = lazy(() =>
  import("./components/files/CodeEditor").then(({ CodeEditor: component }) => ({
    default: component,
  })),
);

export function ProjectDetailPage() {
  const { projectId } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const workspaceId = useWorkspaceId();
  const navigate = useNavigate();
  const { tokens } = useAuth();
  const accessToken = tokens?.access_token ?? null;

  const { sandbox, isLoading, error } = useSandbox(accessToken, projectId ?? null, true);
  const sandboxStatus = typeof sandbox?.status === "string" ? sandbox.status : null;
  const [selectedFile, setSelectedFile] = useState<{ name: string; path: string } | null>(null);
  const panelParam = searchParams.get("panel");
  const activePanel = panelParam === "agent" || panelParam === "git" ? panelParam : "files";
  const workspaceLabel = workspaceId || "workspace";
  const projectLabel = projectId || "project";

  if (!workspaceId || !projectId) {
    return <div>Invalid project</div>;
  }

  function selectPanel(panel: "files" | "agent" | "git") {
    const nextParams = new URLSearchParams(searchParams);
    if (panel === "agent") nextParams.set("panel", "agent");
    else if (panel === "git") nextParams.set("panel", "git");
    else nextParams.delete("panel");
    setSearchParams(nextParams, { replace: true });
  }

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden bg-forge-bg text-forge-text">
      <header className="flex shrink-0 items-center justify-between border-b border-[var(--border)] bg-forge-bg px-4 py-3">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate(`/workspace/projects?workspace=${encodeURIComponent(workspaceId)}`)}
            aria-label="Back to projects"
            className="grid size-8 place-items-center rounded-md border border-[var(--border)] bg-forge-panel text-forge-muted transition hover:bg-[var(--surface-hover)] hover:text-forge-text"
          >
            <ArrowLeft size={15} />
          </button>
          <div className="flex items-center gap-2">
            <div className="grid size-7 place-items-center rounded-md bg-forge-accent text-[11px] font-bold text-[var(--primary-foreground)]">F</div>
            <div className="text-sm font-semibold text-forge-text">{projectLabel}</div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button
            type="button"
            aria-pressed={activePanel === "agent"}
            onClick={() => selectPanel(activePanel === "agent" ? "files" : "agent")}
            className={`inline-flex h-8 items-center gap-2 rounded-md border px-3 text-xs font-semibold transition ${activePanel === "agent" ? "border-forge-accent/40 bg-forge-accent/[0.1] text-forge-accent" : "border-[var(--border)] text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"}`}
          >
            <Bot size={14} /> Agent
          </button>
          <button
            type="button"
            aria-pressed={activePanel === "git"}
            onClick={() => selectPanel(activePanel === "git" ? "files" : "git")}
            className={`inline-flex h-8 items-center gap-2 rounded-md border px-3 text-xs font-semibold transition ${activePanel === "git" ? "border-forge-accent/40 bg-forge-accent/[0.1] text-forge-accent" : "border-[var(--border)] text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"}`}
          >
            <FolderGit2 size={14} /> Git
          </button>
          <div className="rounded-md border border-[var(--border)] bg-forge-panel px-3 py-1.5 text-[11px] font-medium text-forge-muted">
            {workspaceLabel}
          </div>
          {sandboxStatus && <SandboxStatus status={sandboxStatus} />}
        </div>
      </header>

      <div className="relative flex min-h-0 flex-1 overflow-hidden">
        <aside className="flex w-[270px] shrink-0 flex-col border-r border-[var(--border)] bg-forge-panel p-0">
          <div className="flex h-12 shrink-0 items-center gap-2 border-b border-[var(--border)] bg-[var(--surface-subtle)] px-4 text-xs font-semibold text-forge-text">
            <FileText size={14} className="text-forge-muted" /> Files
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto px-3 py-3">
            {sandboxStatus === "RUNNING" ? (
              <FileExplorer
                key={sandbox?.id ?? ""}
                accessToken={accessToken ?? ""}
                sandboxId={sandbox?.id ?? ""}
                selectedPath={selectedFile?.path}
                onSelect={(file) => setSelectedFile({ name: file.name, path: file.path })}
                onPathChanged={(oldPath, newPath) => {
                  setSelectedFile((current) => {
                    if (!current || (current.path !== oldPath && !current.path.startsWith(`${oldPath}/`))) return current;
                    const path = `${newPath}${current.path.slice(oldPath.length)}`;
                    return { name: path.split("/").at(-1) ?? current.name, path };
                  });
                }}
                onPathDeleted={(path) => {
                  setSelectedFile((current) =>
                    current && (current.path === path || current.path.startsWith(`${path}/`))
                      ? null
                      : current,
                  );
                }}
              />
            ) : (
              <div className="text-xs text-forge-muted">
                {isLoading && "Loading files..."}
                {!isLoading && !sandbox && "Environment is being set up..."}
                {sandboxStatus && `Environment ${sandboxStatus.toLowerCase()}`}
              </div>
            )}
          </div>
        </aside>

        <main className="flex min-h-0 min-w-0 flex-1 flex-col bg-forge-bg">
          <div className="flex min-h-0 min-w-0 flex-1 bg-forge-bg p-0">
            {sandboxStatus === "RUNNING" ? (
              selectedFile ? (
                <Suspense
                  fallback={<div className="p-4 text-sm text-forge-muted">Loading editor...</div>}
                >
                  <CodeEditor
                    accessToken={accessToken}
                    sandboxId={sandbox?.id ?? null}
                    filePath={selectedFile.path}
                    fileName={selectedFile.name}
                  />
                </Suspense>
              ) : (
                <div className="flex h-full min-w-0 flex-1 flex-col items-center justify-center gap-3 text-center text-sm text-forge-muted">
                  <Code2 size={28} strokeWidth={1.5} />
                  <span>Select a file to view its contents</span>
                </div>
              )
            ) : (
              <div className="flex h-full items-center justify-center">
                <div className="text-center">
                  {isLoading && (
                    <>
                      <Loader2 size={32} className="mx-auto animate-spin text-forge-accent" />
                      <p className="mt-4 text-sm text-forge-muted">Loading environment...</p>
                    </>
                  )}
                  {!isLoading && error && (
                    <>
                      <div className="text-forge-signal">Environment setup failed</div>
                      <p className="mt-2 text-xs text-forge-muted">Please try again</p>
                    </>
                  )}
                  {!isLoading && sandbox && sandboxStatus && (
                    <>
                      <div className="text-forge-accent">
                        <SandboxStatus status={sandboxStatus} showLabel={false} className="justify-center" />
                      </div>
                      <p className="mt-4 text-sm text-forge-muted">
                        Environment is {sandboxStatus.toLowerCase()}
                      </p>
                      {sandboxStatus === "FAILED" && sandbox.last_error && (
                        <p className="mt-2 text-xs text-forge-signal">{sandbox.last_error}</p>
                      )}
                    </>
                  )}
                  {!isLoading && sandbox && !sandboxStatus && (
                    <div className="text-forge-signal">Environment status is unavailable</div>
                  )}
                </div>
              </div>
            )}
          </div>
        </main>
        {activePanel === "agent" && (
          <ProjectSidePanel
            title="Agent"
            icon={<Bot size={15} className="text-forge-accent" />}
            onClose={() => selectPanel("files")}
          >
            <AgentPanel
              projectName={projectLabel}
              selectedFile={selectedFile?.name ?? null}
              accessToken={accessToken}
              sandboxId={sandbox?.id ?? null}
              sandboxStatus={sandboxStatus}
            />
          </ProjectSidePanel>
        )}
        {activePanel === "git" && (
          <ProjectSidePanel
            title="Git"
            icon={<FolderGit2 size={15} className="text-forge-accent" />}
            onClose={() => selectPanel("files")}
          >
            <GitPanel
              accessToken={accessToken}
              sandboxId={sandbox?.id ?? null}
              projectName={projectLabel}
            />
          </ProjectSidePanel>
        )}
      </div>
    </div>
  );
}

