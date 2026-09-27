import { ChevronDown, ChevronRight, FileIcon, FolderGit2 } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { listFiles, type FileEntry } from "../../../api/sandbox.api";

export function TreeNode({
  file,
  sandboxId,
  accessToken,
  onSelect,
  selectedPath,
}: {
  file: FileEntry;
  sandboxId: string;
  accessToken: string;
  onSelect?: (file: FileEntry) => void;
  selectedPath?: string | null;
}) {
  const [expanded, setExpanded] = useState(true);

  const { data, isLoading, error } = useQuery({
    queryKey: ["sandbox-files", sandboxId, file.path],
    queryFn: () => listFiles(accessToken, sandboxId, file.path),
    enabled: !!sandboxId && !!accessToken && file.is_directory && expanded,
    staleTime: 30_000,
  });

  const children = data?.files ?? [];

  if (!file.is_directory) {
    return (
      <button
        type="button"
        onClick={() => onSelect?.(file)}
        aria-current={selectedPath === file.path ? "true" : undefined}
        className={`flex w-full items-center gap-2 rounded-md px-2 py-1 text-left transition hover:bg-[var(--surface-hover)] hover:text-forge-text ${selectedPath === file.path ? "bg-forge-accent/15 text-forge-text" : "text-forge-soft"}`}
      >
        <FileIcon size={14} className="shrink-0 text-forge-muted" />
        <span className="truncate">{file.name}</span>
      </button>
    );
  }

  return (
    <div className="space-y-1">
      <button
        type="button"
        className="flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-forge-soft transition hover:bg-[var(--surface-hover)] hover:text-forge-text"
        onClick={() => setExpanded((value) => !value)}
      >
        {expanded ? <ChevronDown size={12} className="text-forge-muted" /> : <ChevronRight size={12} className="text-forge-muted" />}
        <FolderGit2 size={14} className="shrink-0 text-forge-muted" />
        <span className="truncate font-medium">{file.name}</span>
      </button>

      {expanded && (
        <div className="ml-4 space-y-1 border-l border-[var(--border)] pl-2">
          {isLoading && <div className="px-2 py-1 text-xs text-forge-muted">Loading files...</div>}
          {!isLoading && error && <div className="px-2 py-1 text-xs text-forge-signal">Unable to load files</div>}
          {!isLoading && !error && children.map((child) => (
            <TreeNode key={child.path} file={child} sandboxId={sandboxId} accessToken={accessToken} onSelect={onSelect} selectedPath={selectedPath} />
          ))}
        </div>
      )}
    </div>
  );
}
