import { FolderGit2 } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { useAuth } from "../../../../auth/useAuth";
import { listFiles, type FileEntry } from "../../../api/sandbox.api";
import { useSandbox } from "../hooks/useSandbox";
import { TreeNode } from "./TreeNode";

export function FileTree({
  onSelect,
  selectedPath,
}: {
  onSelect?: (file: FileEntry) => void;
  selectedPath?: string | null;
}) {
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
    <div className="space-y-2">
      <div className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm text-forge-text">
        <FolderGit2 size={14} className="text-forge-muted" />
        <span className="font-medium">workspace</span>
      </div>
      <div className="space-y-1 text-sm text-forge-soft">
        {isLoading && <div className="px-2 py-1 text-xs text-forge-muted">Loading files...</div>}
        {!isLoading && error && <div className="px-2 py-1 text-xs text-forge-signal">Unable to load files</div>}
        {!isLoading && !error && files.length === 0 && <div className="px-2 py-1 text-xs text-forge-muted">No files found</div>}
        {!isLoading && !error && files.map((file) => (
          <TreeNode key={file.path} file={file} sandboxId={sandbox?.id ?? ""} accessToken={tokens?.access_token ?? ""} onSelect={onSelect} selectedPath={selectedPath} />
        ))}
      </div>
    </div>
  );
}
