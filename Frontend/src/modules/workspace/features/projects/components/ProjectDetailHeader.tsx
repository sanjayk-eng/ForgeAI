import { ArrowLeft, Bot, FolderGit2, Terminal } from "lucide-react";

interface ProjectDetailHeaderProps {
  workspaceId: string;
  projectLabel: string;
  activePanel: "files" | "agent" | "git" | "terminal";
  sandboxStatus: string | null;
  onBack: () => void;
  onSelectPanel: (panel: "files" | "agent" | "git" | "terminal") => void;
}

export function ProjectDetailHeader({
  workspaceId,
  projectLabel,
  activePanel,
  sandboxStatus,
  onBack,
  onSelectPanel,
}: ProjectDetailHeaderProps) {
  return (
    <header className="flex shrink-0 items-center justify-between border-b border-[var(--border)] bg-forge-bg px-4 py-3">
      <div className="flex items-center gap-3">
        <button
          onClick={onBack}
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
          onClick={() => onSelectPanel(activePanel === "agent" ? "files" : "agent")}
          className={`inline-flex h-8 items-center gap-2 rounded-md border px-3 text-xs font-semibold transition ${activePanel === "agent" ? "border-forge-accent/40 bg-forge-accent/[0.1] text-forge-accent" : "border-[var(--border)] text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"}`}
        >
          <Bot size={14} /> Agent
        </button>
        <button
          type="button"
          aria-pressed={activePanel === "git"}
          onClick={() => onSelectPanel(activePanel === "git" ? "files" : "git")}
          className={`inline-flex h-8 items-center gap-2 rounded-md border px-3 text-xs font-semibold transition ${activePanel === "git" ? "border-forge-accent/40 bg-forge-accent/[0.1] text-forge-accent" : "border-[var(--border)] text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"}`}
        >
          <FolderGit2 size={14} /> Git
        </button>
        <button
          type="button"
          aria-pressed={activePanel === "terminal"}
          onClick={() => onSelectPanel(activePanel === "terminal" ? "files" : "terminal")}
          className={`inline-flex h-8 items-center gap-2 rounded-md border px-3 text-xs font-semibold transition ${activePanel === "terminal" ? "border-forge-accent/40 bg-forge-accent/[0.1] text-forge-accent" : "border-[var(--border)] text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"}`}
        >
          <Terminal size={14} /> Terminal
        </button>
        <div className="rounded-md border border-[var(--border)] bg-forge-panel px-3 py-1.5 text-[11px] font-medium text-forge-muted">
          {workspaceId}
        </div>
        {sandboxStatus && (
          <div className="rounded-md border border-[var(--border)] bg-forge-panel px-2 py-1 text-[10px] font-medium uppercase tracking-[0.12em] text-forge-muted">
            {sandboxStatus}
          </div>
        )}
      </div>
    </header>
  );
}
