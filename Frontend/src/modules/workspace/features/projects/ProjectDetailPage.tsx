import { ArrowLeft, Loader2, Terminal as TerminalIcon, FolderGit2 } from "lucide-react";
import { useNavigate, useParams } from "react-router-dom";
import { useAuth } from "../../../auth/useAuth";
import { useWorkspaceId } from "../../hooks/useWorkspaceId";
import { useSandbox } from "./hooks/useSandbox";
import { SandboxStatus } from "./components/SandboxStatus";

export function ProjectDetailPage() {
  const { projectId } = useParams();
  const workspaceId = useWorkspaceId();
  const navigate = useNavigate();
  const { tokens } = useAuth();
  const accessToken = tokens?.access_token ?? null;

  const { sandbox, isLoading, error } = useSandbox(accessToken, projectId ?? null, true);

  if (!workspaceId || !projectId) {
    return <div>Invalid project</div>;
  }

  return (
    <div className="flex h-screen flex-col bg-[var(--background)]">
      <header className="flex items-center justify-between border-b border-[var(--border)] bg-[var(--surface)] px-6 py-4">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate(`/workspace/projects?workspace=${encodeURIComponent(workspaceId)}`)}
            className="grid size-9 place-items-center border border-[var(--border)] text-forge-muted transition hover:border-forge-accent/40 hover:text-forge-text"
          >
            <ArrowLeft size={16} />
          </button>
          <div>
            <h1 className="text-lg font-bold text-forge-text">Project Environment</h1>
            <p className="text-xs text-forge-muted">/{projectId.substring(0, 8)}</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          {sandbox && <SandboxStatus status={sandbox.status} />}
          {isLoading && (
            <div className="flex items-center gap-2 text-xs text-forge-muted">
              <Loader2 size={14} className="animate-spin" />
              Checking environment...
            </div>
          )}
        </div>
      </header>

      <div className="flex flex-1 overflow-hidden">
        <aside className="w-64 border-r border-[var(--border)] bg-[var(--surface)] p-4">
          <h2 className="mb-3 text-xs font-bold uppercase tracking-wider text-forge-soft">
            File Explorer
          </h2>
          {sandbox?.status === "RUNNING" ? (
            <FileTree />
          ) : (
            <div className="text-xs text-forge-muted">
              {isLoading && "Loading files..."}
              {!isLoading && !sandbox && "Environment is being set up..."}
              {sandbox && `Environment ${sandbox.status.toLowerCase()}`}
            </div>
          )}
        </aside>

        <main className="flex flex-1 flex-col">
          <div className="flex-1 bg-[var(--background)] p-6">
            {sandbox?.status === "RUNNING" ? (
              <TerminalView />
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
                  {!isLoading && sandbox && (
                    <>
                      <div className="text-forge-accent">
                        <SandboxStatus status={sandbox.status} showLabel={false} className="justify-center" />
                      </div>
                      <p className="mt-4 text-sm text-forge-muted">
                        Environment is {sandbox.status.toLowerCase()}
                      </p>
                      {sandbox.status === "FAILED" && sandbox.last_error && (
                        <p className="mt-2 text-xs text-forge-signal">{sandbox.last_error}</p>
                      )}
                    </>
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

function FileTree() {
  return (
    <div className="space-y-1">
      <div className="flex items-center gap-2 rounded px-2 py-1.5 text-xs text-forge-text hover:bg-[var(--surface-hover)]">
        <FolderGit2 size={14} />
        <span>/workspace</span>
      </div>
      <div className="ml-4 space-y-1 text-xs text-forge-muted">
        <div className="px-2 py-1">Loading files...</div>
      </div>
    </div>
  );
}

function TerminalView() {
  return (
    <div className="flex h-full flex-col rounded border border-[var(--border)] bg-black/90 font-mono text-sm">
      <div className="flex items-center gap-2 border-b border-gray-700/50 bg-gray-800/50 px-4 py-2">
        <TerminalIcon size={14} className="text-emerald-400" />
        <span className="text-xs text-gray-300">Terminal</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        <div className="text-gray-400">
          <div className="mb-2">
            <span className="text-emerald-400">✓</span> Container ready
          </div>
          <div className="mb-2">
            <span className="text-emerald-400">✓</span> Workspace mounted at /workspace
          </div>
          <div className="mb-4">
            <span className="text-emerald-400">✓</span> Repository cloned
          </div>
          <div className="flex gap-2">
            <span className="text-sky-400">$</span>
            <span className="animate-pulse">_</span>
          </div>
        </div>
      </div>
    </div>
  );
}
