import { useState } from "react";
import { FolderKanban, ChevronDown, Check } from "lucide-react";
import { Dropdown } from "../../../shared/ui/Dropdown";
import type { Workspace } from "../types/workspace.types";

type WorkspaceSelectorProps = {
  workspaces: Workspace[];
  selected?: Workspace;
  onSelect: (workspace: Workspace) => void;
};

export function WorkspaceSelector({ workspaces, selected, onSelect }: WorkspaceSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);

  if (workspaces.length === 0) {
    return (
      <div className="flex items-center gap-2 text-forge-soft">
        <FolderKanban size={15} />
        <span className="text-sm font-bold text-forge-muted">No workspace yet</span>
      </div>
    );
  }

  const handleSelect = (workspace: Workspace) => {
    onSelect(workspace);
    setIsOpen(false);
  };

  return (
    <div className="relative">
      <div className="flex items-center gap-2 text-forge-soft">
        <FolderKanban size={15} />
        
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="group flex max-w-[150px] items-center gap-2 rounded-md px-2 py-1.5 text-sm font-bold text-forge-text transition hover:bg-white/[0.06] sm:max-w-[220px]"
          aria-label="Select workspace"
        >
          <span className="truncate">{selected?.name ?? "Select workspace"}</span>
          <ChevronDown 
            size={14} 
            className="shrink-0 transition-transform group-hover:translate-y-0.5" 
          />
        </button>
      </div>

      <Dropdown 
        isOpen={isOpen} 
        onClose={() => setIsOpen(false)}
        className="w-[min(300px,calc(100vw-32px))]"
        align="left"
      >
        {/* Header */}
        <div className="border-b border-white/10 bg-[#1f2329] px-4 py-3">
          <h3 className="text-xs font-semibold text-white">
            Switch Workspace
          </h3>
          <p className="mt-0.5 text-xs text-gray-400">
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
                className={`flex w-full items-center gap-3 px-4 py-2.5 text-left transition ${
                  isSelected 
                    ? "bg-white/10 text-white" 
                    : "text-gray-300 hover:bg-white/5 hover:text-white"
                }`}
              >
                {/* Icon */}
                <div className={`grid size-8 shrink-0 place-items-center rounded-lg border transition ${
                  isSelected
                    ? "border-forge-accent/50 bg-forge-accent/15 text-forge-accent"
                    : "border-white/15 bg-white/5 text-gray-400"
                }`}>
                  <FolderKanban size={15} />
                </div>

                {/* Content */}
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">
                    {workspace.name}
                  </p>
                  <p className="text-xs text-gray-400">
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
      </Dropdown>
    </div>
  );
}
