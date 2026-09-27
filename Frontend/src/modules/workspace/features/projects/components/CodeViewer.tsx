import { useQuery } from "@tanstack/react-query";
import { Code2, FileCode2 } from "lucide-react";
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
      <div className="flex h-full flex-col items-center justify-center gap-3 text-sm text-forge-muted">
        <Code2 size={28} strokeWidth={1.5} />
        <span>Select a file to view its contents</span>
      </div>
    );
  }

  if (isLoading) {
    return <div className="p-4 text-sm text-forge-muted">Loading file...</div>;
  }

  if (error || !data) {
    return <div className="p-4 text-sm text-forge-signal">Unable to load file content</div>;
  }

  const lines = (data.content ?? "").split("\n");

  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col bg-forge-bg text-forge-text">
      <div className="flex min-h-12 shrink-0 items-center gap-3 border-b border-[var(--border)] bg-forge-panel px-4 text-sm">
        <FileCode2 size={16} className="shrink-0 text-forge-muted" />
        <span className="truncate font-medium">{fileName ?? filePath}</span>
        <span className="hidden truncate text-xs text-forge-muted sm:block">{filePath}</span>
      </div>
      <div className="min-h-0 flex-1 overflow-x-auto overflow-y-auto overscroll-contain bg-forge-bg py-4 font-mono text-[13px] leading-6 selection:bg-forge-accent/25">
        <div className="min-w-max pr-8">
          {lines.map((line, index) => (
            <div key={index} className="flex min-h-6 whitespace-pre">
              <span className="sticky left-0 w-14 shrink-0 select-none border-r border-[var(--border)] bg-forge-bg pr-3 text-right text-forge-muted">
                {index + 1}
              </span>
              <span className="whitespace-pre pl-4 text-forge-text">{line || " "}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
