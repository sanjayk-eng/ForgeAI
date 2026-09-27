import { ArrowLeft, Loader2 } from "lucide-react";
import { useNavigate, useParams } from "react-router-dom";
import { useAuth } from "../../../auth/useAuth";
import { useWorkspaceId } from "../../hooks/useWorkspaceId";
import { useSandbox } from "./hooks/useSandbox";
import { SandboxStatus } from "./components/SandboxStatus";
import { FileTree } from "./components/FileTree";
import { CodeViewer } from "./components/CodeViewer";
import { useState } from "react";

export function ProjectDetailPage() {
  const { projectId } = useParams();
  const workspaceId = useWorkspaceId();
  const navigate = useNavigate();
  const { tokens } = useAuth();
  const accessToken = tokens?.access_token ?? null;

  const { sandbox, isLoading, error } = useSandbox(accessToken, projectId ?? null, true);
  const sandboxStatus = typeof sandbox?.status === "string" ? sandbox.status : null;
  const [selectedFile, setSelectedFile] = useState<{ name: string; path: string } | null>(null);
  const workspaceLabel = workspaceId || "workspace";
  const projectLabel = projectId || "project";

  if (!workspaceId || !projectId) {
    return <div>Invalid project</div>;
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
          <div className="rounded-md border border-[var(--border)] bg-forge-panel px-3 py-1.5 text-[11px] font-medium text-forge-muted">
            {workspaceLabel}
          </div>
          {sandboxStatus && <SandboxStatus status={sandboxStatus} />}
        </div>
      </header>

      <div className="flex min-h-0 flex-1 overflow-hidden">
        <aside className="flex w-[270px] shrink-0 flex-col border-r border-[var(--border)] bg-forge-panel p-0">
          <div className="shrink-0 border-b border-[var(--border)] bg-[var(--surface-subtle)] px-4 py-3 text-xs font-semibold text-forge-text">
            Files
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto px-3 py-3">
            {sandboxStatus === "RUNNING" ? (
              <FileTree
                selectedPath={selectedFile?.path}
                onSelect={(file) => setSelectedFile({ name: file.name, path: file.path })}
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
          <div className="flex min-h-0 flex-1 bg-forge-bg p-0">
            {sandboxStatus === "RUNNING" ? (
              <CodeViewer sandboxId={sandbox?.id ?? null} filePath={selectedFile?.path ?? null} fileName={selectedFile?.name ?? null} />
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
      </div>
    </div>
  );
}

