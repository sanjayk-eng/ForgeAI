import { Code2, Loader2 } from "lucide-react";
import { lazy, Suspense } from "react";
import type { SequencedProjectRealtimeEvent } from "../../../api/project-realtime.types";
import type { Sandbox, SandboxStatus as SandboxStatusType } from "../../../api/sandbox.api";
import { SandboxStatus } from "./SandboxStatus";

const CodeEditor = lazy(() =>
  import("./files/CodeEditor").then(({ CodeEditor: component }) => ({
    default: component,
  })),
);
const GitPanel = lazy(() =>
  import("./GitPanel").then(({ GitPanel: component }) => ({
    default: component,
  })),
);
const ManualTerminal = lazy(() =>
  import("./ManualTerminal").then(({ ManualTerminal: component }) => ({
    default: component,
  })),
);

interface ProjectDetailMainContentProps {
  activePanel: "files" | "agent" | "git" | "terminal";
  accessToken: string | null;
  sandbox: Sandbox | undefined;
  sandboxStatus: SandboxStatusType | null;
  isLoading: boolean;
  error: Error | null;
  projectLabel: string;
  selectedFile: { name: string; path: string } | null;
  realtimeEvents: SequencedProjectRealtimeEvent[];
  resyncVersion: number;
  terminalVisited: boolean;
  projectId: string;
}

export function ProjectDetailMainContent({
  activePanel,
  accessToken,
  sandbox,
  sandboxStatus,
  isLoading,
  error,
  projectLabel,
  selectedFile,
  realtimeEvents,
  resyncVersion,
  terminalVisited,
  projectId,
}: ProjectDetailMainContentProps) {
  return (
    <>
      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        <div className="flex min-h-0 min-w-0 flex-1 bg-forge-bg p-0">
          {activePanel === "git" ? (
            <Suspense fallback={<div className="p-4 text-sm text-forge-muted">Loading Git review...</div>}>
              <GitPanel
                accessToken={accessToken}
                sandboxId={sandbox?.id ?? null}
                sandboxStatus={sandboxStatus}
                projectName={projectLabel}
              />
            </Suspense>
          ) : sandboxStatus === "RUNNING" ? (
            selectedFile ? (
              <Suspense fallback={<div className="p-4 text-sm text-forge-muted">Loading editor...</div>}>
                <CodeEditor
                  accessToken={accessToken}
                  sandboxId={sandbox?.id ?? null}
                  filePath={selectedFile.path}
                  fileName={selectedFile.name}
                  realtimeEvents={realtimeEvents}
                  resyncVersion={resyncVersion}
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
        {terminalVisited && (
          <div
            className={`min-h-0 shrink-0 border-t border-[var(--border)] ${activePanel === "terminal" ? "flex" : "hidden"}`}
            style={{ height: "min(36vh, 320px)", minHeight: 180 }}
          >
            <Suspense fallback={<div className="p-4 text-sm text-forge-muted">Opening terminal...</div>}>
              <ManualTerminal key={projectId} projectId={projectId} accessToken={accessToken} />
            </Suspense>
          </div>
        )}
      </div>
    </>
  );
}
