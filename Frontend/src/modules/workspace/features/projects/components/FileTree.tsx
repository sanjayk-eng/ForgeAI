import { FolderGit2 } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { useAuth } from "../../../../auth/useAuth";
import { listFiles, type FileEntry } from "../../../api/sandbox.api";
import { useSandbox } from "../hooks/useSandbox";
import { TreeNode } from "./TreeNode";

export function FileTree({ onSelect }: { onSelect?: (file: FileEntry) => void }) {
  const { tokens } = useAuth();
  const { projectId } = useParams();
  const { sandbox } = useSandbox(tokens?.access_token ?? null, projectId ?? null, false);

  const { data, isLoading, error } = useQuery({
    queryKey: ["sandbox-files", sandbox?.id, "/workspace"],
    queryFn: () => {
      if (!sandbox?.id || !tokens?.access_token) {
        throw new Error("Missing sandbox id or access token");
      }
      return listFiles(tokens.access_token, sandbox.id, "/workspace");
    },
    enabled: !!sandbox?.id && !!tokens?.access_token && sandbox.status === "RUNNING",
    staleTime: 30_000,
  });

  const files = data?.files ?? [];

  return (
    <div className="space-y-1">
      <div className="flex items-center gap-2 rounded px-2 py-1.5 text-xs text-forge-text hover:bg-[var(--surface-hover)]">
        <FolderGit2 size={14} />
        <span>/workspace</span>
      </div>
      <div className="ml-4 space-y-1 text-xs text-forge-muted">
        {isLoading && <div className="px-2 py-1">Loading files...</div>}
        {!isLoading && error && <div className="px-2 py-1 text-forge-signal">Unable to load files</div>}
        {!isLoading && !error && files.length === 0 && <div className="px-2 py-1">No files found</div>}
        {!isLoading && !error && files.map((file) => (
          <TreeNode key={file.path} file={file} sandboxId={sandbox?.id ?? ""} accessToken={tokens?.access_token ?? ""} onSelect={onSelect} />
        ))}
      </div>
    </div>
  );
}
