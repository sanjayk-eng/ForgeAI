import { DiffEditor } from "@monaco-editor/react";
import { useQuery } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { useState } from "react";
import { useTheme } from "../../../../../shared/ui/themeContextStore";
import { getGitDiff, type GitDiffEntry, type GitStatusPayload } from "../../../api/git.api";
import { readFile } from "../../../api/sandbox-files.api";
import { getFileLanguage } from "./files/fileLanguage";

interface GitDiffPreviewProps {
  accessToken: string | null;
  sandboxId: string | null;
  status: GitStatusPayload | undefined;
  files: GitDiffEntry[];
}

interface ChangedFile {
  path: string;
  status: string;
}

export function GitDiffPreview({ accessToken, sandboxId, status, files }: GitDiffPreviewProps) {
  const { resolvedTheme } = useTheme();
  const [requestedPath, setRequestedPath] = useState<string | null>(null);
  const changedFiles = collectChangedFiles(status, files);
  const selectedPath = changedFiles.some(({ path }) => path === requestedPath)
    ? requestedPath
    : changedFiles[0]?.path ?? null;
  const selectedFile = changedFiles.find(({ path }) => path === selectedPath);

  const selectedDiff = useQuery({
    queryKey: ["git-file-diff", sandboxId, selectedPath],
    queryFn: async () => {
      if (!accessToken || !sandboxId || !selectedPath) {
        throw new Error("Missing Git review context");
      }
      const response = await getGitDiff(accessToken, sandboxId, selectedPath);
      const entry = response.files.find((file) => file.path === selectedPath || file.old_path === selectedPath);
      const original = entry?.original_content ?? "";
      const modified = entry?.status === "deleted"
        ? ""
        : (await readFile(accessToken, sandboxId, `/workspace/${selectedPath}`)).content;
      return { original, modified };
    },
    enabled: Boolean(accessToken && sandboxId && selectedPath),
  });

  return (
    <div className="flex min-h-0 flex-1 overflow-hidden border border-[var(--border)] bg-forge-bg">
      <nav aria-label="Changed files" className="w-56 shrink-0 overflow-y-auto border-r border-[var(--border)] bg-forge-panel">
        <div className="flex h-9 items-center justify-between border-b border-[var(--border)] px-3 text-[10px] font-semibold uppercase text-forge-muted">
          <span>Changes</span>
          <span>{changedFiles.length}</span>
        </div>
        {changedFiles.map((file) => (
          <button
            key={file.path}
            type="button"
            aria-current={file.path === selectedPath ? "true" : undefined}
            onClick={() => setRequestedPath(file.path)}
            className={`flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-xs hover:bg-[var(--surface-hover)] ${file.path === selectedPath ? "bg-[var(--surface-hover)] text-forge-text" : "text-forge-muted"}`}
          >
            <span className="min-w-0 truncate font-mono">{file.path}</span>
            <span className={`shrink-0 text-[9px] font-bold uppercase ${fileStatusClass(file.status)}`}>
              {fileStatusLabel(file.status)}
            </span>
          </button>
        ))}
        {changedFiles.length === 0 && (
          <p className="px-3 py-4 text-xs text-forge-muted">No changed files</p>
        )}
      </nav>

      <section className="flex min-h-0 min-w-0 flex-1 flex-col">
        {selectedFile ? (
          <>
            <header className="flex h-9 shrink-0 items-center justify-between gap-3 border-b border-[var(--border)] px-3">
              <span className="truncate font-mono text-xs text-forge-text">{selectedFile.path}</span>
              <span className={`shrink-0 text-[10px] uppercase ${fileStatusClass(selectedFile.status)}`}>
                {selectedFile.status}
              </span>
            </header>
            {selectedDiff.isLoading ? (
              <div className="flex min-h-0 flex-1 items-center justify-center gap-2 text-xs text-forge-muted">
                <Loader2 size={14} className="animate-spin" /> Loading diff
              </div>
            ) : selectedDiff.isError ? (
              <div className="p-4 text-xs text-forge-signal">
                {selectedDiff.error instanceof Error ? selectedDiff.error.message : "Could not load file diff."}
              </div>
            ) : selectedDiff.data ? (
              <DiffEditor
                width="100%"
                height="100%"
                original={selectedDiff.data.original}
                modified={selectedDiff.data.modified}
                language={getFileLanguage(selectedFile.path)}
                theme={resolvedTheme === "dark" ? "forge-dark" : "forge-light"}
                options={{
                  automaticLayout: true,
                  readOnly: true,
                  minimap: { enabled: false },
                  scrollBeyondLastLine: false,
                  renderSideBySide: true,
                }}
              />
            ) : null}
          </>
        ) : (
          <div className="flex flex-1 items-center justify-center text-sm text-forge-muted">
            Select a changed file to review its diff
          </div>
        )}
      </section>
    </div>
  );
}

function collectChangedFiles(status: GitStatusPayload | undefined, diffs: GitDiffEntry[]): ChangedFile[] {
  const files = new Map<string, string>();
  for (const file of diffs) files.set(file.path, file.status);
  for (const path of status?.staged ?? []) if (!files.has(path)) files.set(path, "staged");
  for (const path of status?.modified ?? []) if (!files.has(path)) files.set(path, "modified");
  for (const path of status?.untracked ?? []) files.set(path, "untracked");
  return [...files].map(([path, fileStatus]) => ({ path, status: fileStatus }))
    .sort((left, right) => left.path.localeCompare(right.path));
}

function fileStatusLabel(status: string) {
  if (status === "untracked" || status === "added") return "A";
  if (status === "deleted") return "D";
  if (status === "renamed") return "R";
  return status === "staged" ? "S" : "M";
}

function fileStatusClass(status: string) {
  if (status === "untracked" || status === "added") return "text-emerald-400";
  if (status === "deleted") return "text-rose-400";
  if (status === "renamed") return "text-forge-accent";
  return "text-amber-400";
}