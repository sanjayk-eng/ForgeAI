import { Bot, FileText } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useQueryClient, type QueryClient } from "@tanstack/react-query";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { useAuth } from "../../../auth/useAuth";
import { getGitStatus } from "../../api/git.api";
import type { ProjectRealtimeEvent, SequencedProjectRealtimeEvent } from "../../api/project-realtime.types";
import { useWorkspaceId } from "../../hooks/useWorkspaceId";
import { useSandbox } from "./hooks/useSandbox";
import { useProjectRealtime } from "./hooks/useProjectRealtime";
import { FileExplorer } from "./components/files/FileExplorer";
import { sandboxFileKeys, WORKSPACE_ROOT } from "./components/files/fileTree";
import { AgentPanel } from "./components/AgentPanel";
import { ProjectSidePanel } from "./components/ProjectSidePanel";
import { ProjectDetailHeader } from "./components/ProjectDetailHeader";
import { ProjectDetailMainContent } from "./components/ProjectDetailMainContent";

export function ProjectDetailPage() {
  const { projectId } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const workspaceId = useWorkspaceId();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { tokens } = useAuth();
  const accessToken = tokens?.access_token ?? null;

  const { sandbox, isLoading, error } = useSandbox(accessToken, projectId ?? null, true);
  const sandboxStatus = typeof sandbox?.status === "string" ? sandbox.status : null;
  const [selectedFile, setSelectedFile] = useState<{ name: string; path: string } | null>(null);
  const [changedPaths, setChangedPaths] = useState<string[]>([]);
  const [realtimeEvents, setRealtimeEvents] = useState<SequencedProjectRealtimeEvent[]>([]);
  const [resyncVersion, setResyncVersion] = useState(0);
  const eventSequence = useRef(0);
  const gitRefreshCooldownRef = useRef(0);
  const gitStatusRefreshCooldownRef = useRef(0);
  const panelParam = searchParams.get("panel");
  const activePanel = panelParam === "agent" || panelParam === "git" || panelParam === "terminal" ? panelParam : "files";
  const [terminalVisited, setTerminalVisited] = useState(activePanel === "terminal");
  const workspaceLabel = workspaceId || "workspace";
  const projectLabel = projectId || "project";

  useEffect(() => {
    eventSequence.current = 0;
    setRealtimeEvents([]);
    setChangedPaths([]);
    setSelectedFile(null);
  }, [projectId]);

  function refreshGitStatus(force = false) {
    if (!accessToken || !sandbox?.id) return;
    const now = Date.now();
    if (!force && now - gitStatusRefreshCooldownRef.current < 1500) return;
    gitStatusRefreshCooldownRef.current = now;
    void queryClient.fetchQuery({
      queryKey: ["git-status", sandbox.id],
      queryFn: () => getGitStatus(accessToken, sandbox.id),
      staleTime: 0,
    }).then(({ status }) => {
      const paths = [...(status.staged ?? []), ...(status.modified ?? []), ...(status.untracked ?? [])]
        .map(toWorkspacePath);
      setChangedPaths([...new Set(paths)]);
    }).catch(() => undefined);
  }

  function refreshGitPanel(force = false) {
    if (!sandbox?.id || activePanel !== "git") return;
    const now = Date.now();
    if (!force && now - gitRefreshCooldownRef.current < 1500) return;
    gitRefreshCooldownRef.current = now;
    invalidateGitDiffs(queryClient, sandbox.id);
    refreshGitStatus(force);
  }

  function resyncProject() {
    setResyncVersion((current) => current + 1);
    if (projectId) void queryClient.invalidateQueries({ queryKey: ["sandbox", projectId] });
    if (!sandbox?.id) return;
    if (activePanel === "git") {
      refreshGitPanel(true);
    }
  }

  function handleRealtimeEvent(event: ProjectRealtimeEvent) {
    if (event.project_id !== projectId) return;
    const sequence = ++eventSequence.current;
    setRealtimeEvents((current) => [...current.slice(-255), { sequence, event }]);
    if (sequence > 0 && sequence % 256 === 0) setResyncVersion((current) => current + 1);

    if (event.event === "sync.required") {
      resyncProject();
      return;
    }

    if (event.event.startsWith("file.")) {
      const path = event.path ? toWorkspacePath(event.path) : "";
      const oldPath = event.old_path ? toWorkspacePath(event.old_path) : "";
      if (sandbox?.id && path) {
        if (path !== selectedFile?.path) {
          void queryClient.invalidateQueries({
            queryKey: sandboxFileKeys.content(sandbox.id, path),
            exact: true,
          });
        }
        if (activePanel === "git") {
          refreshGitPanel();
        }
      }
      if (sandbox?.id && oldPath && oldPath !== selectedFile?.path) {
        void queryClient.invalidateQueries({
          queryKey: sandboxFileKeys.content(sandbox.id, oldPath),
          exact: true,
        });
      }

      if (event.event === "file.deleted") {
        setChangedPaths((current) => current.filter((currentPath) => !isPathWithin(currentPath, path)));
        setSelectedFile((current) => current && isPathWithin(current.path, path) ? null : current);
      } else if (event.event === "file.renamed" && oldPath) {
        setChangedPaths((current) => [...new Set([
          ...current.filter((currentPath) => !isPathWithin(currentPath, oldPath)),
          path,
        ])]);
        setSelectedFile((current) => {
          if (!current || !isPathWithin(current.path, oldPath)) return current;
          const newPath = `${path}${current.path.slice(oldPath.length)}`;
          return { name: newPath.split("/").at(-1) ?? current.name, path: newPath };
        });
      } else if (path) {
        setChangedPaths((current) => current.includes(path) ? current : [...current, path]);
      }
      return;
    }

    if (event.event === "git.status.changed") {
      refreshGitStatus();
      if (activePanel === "git" && sandbox?.id) {
        refreshGitPanel();
      }
      return;
    }

    if (event.event === "sandbox.status.changed") {
      void queryClient.invalidateQueries({ queryKey: ["sandbox", projectId] });
    }
  }

  useEffect(() => {
    if (activePanel === "git" && sandbox?.id) {
      refreshGitPanel(true);
    }
  }, [activePanel, sandbox?.id]);

  useProjectRealtime(accessToken, projectId ?? null, handleRealtimeEvent, resyncProject);

  if (!workspaceId || !projectId) {
    return <div>Invalid project</div>;
  }

  function selectPanel(panel: "files" | "agent" | "git" | "terminal") {
    if (panel === "terminal") setTerminalVisited(true);
    const nextParams = new URLSearchParams(searchParams);
    if (panel === "agent") nextParams.set("panel", "agent");
    else if (panel === "git") nextParams.set("panel", "git");
    else if (panel === "terminal") nextParams.set("panel", "terminal");
    else nextParams.delete("panel");
    setSearchParams(nextParams, { replace: true });
  }

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden bg-forge-bg text-forge-text">
      <ProjectDetailHeader
        workspaceId={workspaceLabel}
        projectLabel={projectLabel}
        activePanel={activePanel}
        sandboxStatus={sandboxStatus}
        onBack={() => navigate(`/workspace/projects?workspace=${encodeURIComponent(workspaceId)}`)}
        onSelectPanel={selectPanel}
      />

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
                changedPaths={changedPaths}
                realtimeEvents={realtimeEvents}
                resyncVersion={resyncVersion}
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
          <ProjectDetailMainContent
            activePanel={activePanel}
            accessToken={accessToken}
            sandbox={sandbox}
            sandboxStatus={sandboxStatus}
            isLoading={isLoading}
            error={error}
            projectLabel={projectLabel}
            selectedFile={selectedFile}
            realtimeEvents={realtimeEvents}
            resyncVersion={resyncVersion}
            terminalVisited={terminalVisited}
            projectId={projectId}
          />
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
              realtimeEvents={realtimeEvents}
            />
          </ProjectSidePanel>
        )}
      </div>
    </div>
  );
}

function toWorkspacePath(path: string) {
  const normalized = path.replaceAll("\\", "/").replace(/^\/+/, "");
  return normalized === "workspace" || normalized.startsWith("workspace/")
    ? `/${normalized}`
    : `${WORKSPACE_ROOT}/${normalized}`;
}

function invalidateGitDiffs(queryClient: QueryClient, sandboxId: string) {
  void queryClient.invalidateQueries({ queryKey: ["git-diff", sandboxId] });
  void queryClient.invalidateQueries({ queryKey: ["git-file-diff", sandboxId] });
}

function isPathWithin(candidate: string, parent: string) {
  return candidate === parent || candidate.startsWith(`${parent}/`);
}

