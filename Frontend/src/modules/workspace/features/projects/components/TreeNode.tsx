import { ChevronDown, ChevronRight, FileIcon, FolderGit2 } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { listFiles, type FileEntry } from "../../../api/sandbox.api";

export function TreeNode({
  file,
  sandboxId,
  accessToken,
  onSelect,
}: {
  file: FileEntry;
  sandboxId: string;
  accessToken: string;
  onSelect?: (file: FileEntry) => void;
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
        className="flex w-full items-center gap-2 rounded px-2 py-1 text-left hover:bg-[var(--surface-hover)]"
      >
        <FileIcon size={14} />
        <span className="text-forge-muted">{file.name}</span>
      </button>
    );
  }

  return (
    <div className="space-y-1">
      <button
        type="button"
        className="flex w-full items-center gap-2 rounded px-2 py-1 text-left hover:bg-[var(--surface-hover)]"
        onClick={() => setExpanded((value) => !value)}
      >
        {expanded ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
        <FolderGit2 size={14} className="text-forge-text" />
        <span className="text-forge-text">{file.name}</span>
      </button>

      {expanded && (
        <div className="ml-4 space-y-1 border-l border-[var(--border)] pl-2">
          {isLoading && <div className="px-2 py-1">Loading files...</div>}
          {!isLoading && error && <div className="px-2 py-1 text-forge-signal">Unable to load files</div>}
          {!isLoading && !error && children.map((child) => (
            <TreeNode key={child.path} file={child} sandboxId={sandboxId} accessToken={accessToken} onSelect={onSelect} />
          ))}
        </div>
      )}
    </div>
  );
}
