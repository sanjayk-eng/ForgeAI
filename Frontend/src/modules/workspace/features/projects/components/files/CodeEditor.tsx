import Editor from "@monaco-editor/react";
import { Code2, FileCode2, Save } from "lucide-react";
import { useEffect, useRef } from "react";
import "./monacoSetup";
import { useTheme } from "../../../../../../shared/ui/themeContextStore";
import { getFileLanguage } from "./fileLanguage";
import { useFileEditor } from "./useFileEditor";

export function CodeEditor({
  accessToken,
  sandboxId,
  filePath,
  fileName,
}: {
  accessToken: string | null;
  sandboxId: string | null;
  filePath: string | null;
  fileName: string | null;
}) {
  const { resolvedTheme } = useTheme();
  const editorState = useFileEditor(sandboxId, filePath, accessToken);
  const saveRef = useRef(editorState.save);

  useEffect(() => {
    saveRef.current = editorState.save;
  }, [editorState.save]);

  if (!filePath || !sandboxId) {
    return (
      <div className="flex h-full min-w-0 flex-1 flex-col items-center justify-center gap-3 text-center text-sm text-forge-muted">
        <Code2 size={28} strokeWidth={1.5} />
        <span>Select a file to view its contents</span>
      </div>
    );
  }

  if (editorState.isLoading) {
    return <div className="p-4 text-sm text-forge-muted">Loading file...</div>;
  }

  if (editorState.error) {
    return <div className="p-4 text-sm text-forge-signal">Unable to load or save file</div>;
  }

  return (
    <div className="flex h-full min-h-0 min-w-0 flex-1 flex-col bg-forge-bg text-forge-text">
      <div className="flex min-h-12 shrink-0 items-center justify-between gap-3 border-b border-[var(--border)] bg-forge-panel px-4 text-sm">
        <div className="flex min-w-0 items-center gap-3">
          <FileCode2 size={16} className="shrink-0 text-forge-muted" />
          <span className="truncate font-medium">{fileName ?? filePath}</span>
          <span className="hidden truncate text-xs text-forge-muted sm:block">{filePath}</span>
        </div>
        <button
          type="button"
          onClick={() => void editorState.save()}
          disabled={!editorState.hasUnsavedChanges || editorState.isSaving}
          className="inline-flex items-center gap-2 rounded-md border border-forge-accent/30 bg-forge-accent/[0.08] px-2.5 py-1.5 text-[11px] font-bold uppercase tracking-[0.08em] text-forge-accent disabled:cursor-not-allowed disabled:opacity-50"
        >
          <Save size={12} />
          {editorState.isSaving ? "Saving..." : editorState.hasUnsavedChanges ? "Save" : "Saved"}
        </button>
      </div>
      <div className="min-h-0 flex-1">
        <Editor
          width="100%"
          height="100%"
          path={filePath}
          language={getFileLanguage(filePath)}
          value={editorState.content}
          onChange={(value) => editorState.updateDraft(value ?? "")}
          onMount={(editor, monaco) => {
            editor.addAction({
              id: "forge.save-file",
              label: "Save file",
              keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS],
              run: () => void saveRef.current(),
            });
          }}
          loading={<div className="p-4 text-sm text-forge-muted">Loading editor...</div>}
          options={{
            automaticLayout: true,
            fontSize: 13,
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            tabSize: 2,
          }}
          theme={resolvedTheme === "dark" ? "forge-dark" : "forge-light"}
        />
      </div>
    </div>
  );
}