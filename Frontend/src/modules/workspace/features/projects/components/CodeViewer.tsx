import { useQuery } from "@tanstack/react-query";
import { useAuth } from "../../../../auth/useAuth";
import { readFile } from "../../../api/sandbox.api";

export function CodeViewer({
  sandboxId,
  filePath,
  fileName,
}: {
  sandboxId: string | null;
  filePath: string | null;
  fileName: string | null;
}) {
  const { tokens } = useAuth();

  const { data, isLoading, error } = useQuery({
    queryKey: ["sandbox-file-content", sandboxId, filePath],
    queryFn: () => {
      if (!sandboxId || !filePath || !tokens?.access_token) {
        throw new Error("Missing sandbox, path or access token");
      }
      return readFile(tokens.access_token, sandboxId, filePath);
    },
    enabled: !!sandboxId && !!filePath && !!tokens?.access_token,
    staleTime: 30_000,
  });

  if (!filePath || !sandboxId) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-forge-muted">
        Select a file to view its contents
      </div>
    );
  }

  if (isLoading) {
    return <div className="p-4 text-sm text-forge-muted">Loading file...</div>;
  }

  if (error || !data) {
    return <div className="p-4 text-sm text-forge-signal">Unable to load file content</div>;
  }

  return (
    <div className="flex h-full flex-col bg-[#0b1220] text-[#dbe7ff]">
      <div className="border-b border-gray-700 bg-gray-900/80 px-4 py-2 text-xs text-gray-300">
        {fileName ?? filePath}
      </div>
      <pre className="flex-1 overflow-auto whitespace-pre-wrap p-4 font-mono text-xs leading-6">
        {data.content || ""}
      </pre>
    </div>
  );
}
