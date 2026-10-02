import { useId, useState } from "react";
import { FolderKanban, ChevronDown, Check, Plus } from "lucide-react";
import { Dropdown } from "../../../shared/ui/Dropdown";
import type { Workspace } from "../types/workspace.types";

type WorkspaceSelectorProps = {
  workspaces: Workspace[];
  selected?: Workspace;
  onSelect: (workspace: Workspace) => void;
  onCreate: () => void;
};

export function WorkspaceSelector({ workspaces, selected, onSelect, onCreate }: WorkspaceSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const menuId = useId();

  if (workspaces.length === 0) {
    return <button type="button" className="group flex items-center gap-2.5 rounded-lg border border-dashed border-forge-accent/35 bg-forge-accent/[0.06] px-3 py-2 text-left transition hover:border-forge-accent hover:bg-forge-accent/[0.12] focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent" onClick={onCreate}>
      <span className="grid size-8 place-items-center rounded-md border border-forge-accent/35 bg-forge-accent/[0.12] text-forge-accent"><Plus size={16} /></span>
      <span><span className="block text-[10px] font-extrabold uppercase tracking-[.1em] text-forge-accent">Workspace</span><span className="block text-sm font-bold text-forge-text">Add workspace</span></span>
    </button>;
  }

  const handleSelect = (workspace: Workspace) => {
    onSelect(workspace);
    setIsOpen(false);
  };

  return (
    <div className="relative z-50">
      <div className="flex items-center gap-2 text-forge-soft">
        <span className="grid size-10 place-items-center rounded-lg border border-forge-accent/30 bg-forge-accent/[0.1] font-mono text-sm font-extrabold text-forge-accent">
          {workspaceInitial(selected?.name)}
        </span>
        
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="group flex max-w-[180px] items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm font-bold text-forge-text transition hover:bg-[var(--surface-hover)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent sm:max-w-[240px]"
          aria-label="Select workspace"
          aria-expanded={isOpen}
          aria-controls={menuId}
        >
          <span className="min-w-0"><span className="block truncate text-[10px] font-extrabold uppercase tracking-[.1em] text-forge-muted">Current workspace</span><span className="block truncate text-sm font-bold">{selected?.name ?? "Select workspace"}</span></span>
          <ChevronDown 
            size={14} 
            className="shrink-0 transition-transform group-hover:translate-y-0.5" 
          />
        </button>
      </div>

      <Dropdown 
        isOpen={isOpen} 
        onClose={() => setIsOpen(false)}
        className="w-[min(320px,calc(100vw-32px))] overflow-hidden"
        align="left"
      >
        {/* Header */}
        <div id={menuId} className="border-b border-[var(--border)] bg-forge-card px-4 py-3">
          <h3 className="text-xs font-bold text-forge-text">
            Switch Workspace
          </h3>
          <p className="mt-0.5 text-xs text-forge-muted">
            {workspaces.length} available
          </p>
        </div>

        {/* Workspace List */}
        <div className="max-h-[360px] overflow-y-auto py-1">
          {workspaces.map((workspace) => {
            const isSelected = workspace.id === selected?.id;
            return (
              <button
                key={workspace.id}
                onClick={() => handleSelect(workspace)}
                  role="menuitem"
                  className={`flex w-full items-center gap-3 px-4 py-3 text-left transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-inset focus-visible:outline-forge-accent ${
                  isSelected 
                    ? "bg-forge-accent/[0.1] text-forge-text" 
                    : "text-forge-soft hover:bg-[var(--surface-hover)] hover:text-forge-text"
                }`}
              >
                {/* Icon */}
                <div className={`grid size-8 shrink-0 place-items-center rounded-lg border transition ${
                  isSelected
                    ? "border-forge-accent/50 bg-forge-accent/15 text-forge-accent"
                    : "border-[var(--border)] bg-[var(--surface-subtle)] text-forge-muted"
                }`}>
                  <FolderKanban size={15} />
                </div>

                {/* Content */}
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">
                    {workspace.name}
                  </p>
                  <p className="text-xs text-forge-muted">
                    /{workspace.slug}
                  </p>
                </div>

                {/* Check mark */}
                {isSelected && (
                  <Check size={15} className="shrink-0 text-forge-accent" />
                )}
              </button>
            );
          })}
        </div>
        <button type="button" className="flex w-full items-center gap-3 border-t border-[var(--border)] px-4 py-3 text-left text-sm font-bold text-forge-accent transition hover:bg-forge-accent/[0.08]" onClick={() => { setIsOpen(false); onCreate(); }}>
          <span className="grid size-8 place-items-center rounded-md border border-forge-accent/30 bg-forge-accent/[0.08]"><Plus size={15} /></span>
          Add workspace
        </button>
      </Dropdown>
    </div>
  );
}

function workspaceInitial(name?: string) {
  return name?.trim().slice(0, 1).toUpperCase() || <FolderKanban size={17} />;
}
