import type { GitDiffEntry } from "../../../api/git.api";

export function GitDiffPreview({ files }: { files: GitDiffEntry[] }) {
  return (
    <div className="space-y-2">
      {files.map((file) => (
        <article key={`${file.old_path ?? ""}:${file.path}`} className="overflow-hidden border border-[var(--border)] bg-forge-panel">
          <header className="flex items-center justify-between gap-3 border-b border-[var(--border)] px-2 py-1.5">
            <div className="min-w-0 truncate font-mono text-[10px] text-forge-accent">
              {file.old_path ? `${file.old_path} -> ` : ""}{file.path}
            </div>
            <span className="shrink-0 text-[9px] uppercase text-forge-muted">{file.status}</span>
          </header>
          <pre className="max-h-80 overflow-auto text-[10px] leading-5">
            <code>
              {displayDiffLines(file.content).map(({ text, kind }, index) => (
                <span
                  key={`${index}-${text}`}
                  className={`block min-w-max px-2 ${lineClass(kind)}`}
                >
                  {text || " "}
                </span>
              ))}
            </code>
          </pre>
        </article>
      ))}
    </div>
  );
}

type DiffLine = { text: string; kind: "added" | "removed" | "hunk" | "context" };

function displayDiffLines(patch: string): DiffLine[] {
  const lines: DiffLine[] = [];
  for (const text of patch.split("\n")) {
    if (
      text.startsWith("a/") || text.startsWith("index ") ||
      text.startsWith("--- ") || text.startsWith("+++ ") ||
      text.startsWith("rename from ") || text.startsWith("rename to ") ||
      text.startsWith("similarity index ") || text.startsWith("new file mode ") ||
      text.startsWith("deleted file mode ") || text.startsWith("diff --git ")
    ) continue;
    if (text.startsWith("@@")) lines.push({ text, kind: "hunk" });
    else if (text.startsWith("+")) lines.push({ text, kind: "added" });
    else if (text.startsWith("-")) lines.push({ text, kind: "removed" });
    else lines.push({ text, kind: "context" });
  }
  return lines;
}

function lineClass(kind: "added" | "removed" | "hunk" | "context") {
  switch (kind) {
    case "added":
      return "bg-emerald-500/10 text-emerald-300";
    case "removed":
      return "bg-rose-500/10 text-rose-300";
    case "hunk":
      return "bg-forge-accent/[0.08] text-forge-accent";
    default:
      return "text-forge-muted";
  }
}