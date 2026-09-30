import { DiffEditor, Editor } from "@monaco-editor/react";
import { Code2, FileCode2, GitCompareArrows, Save } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import type { editor as MonacoEditor } from "monaco-editor";
import type * as Monaco from "monaco-editor";
import "./monacoSetup";
import { useTheme } from "../../../../../../shared/ui/themeContextStore";
import { getGoDefinition, readFile } from "../../../../api/sandbox-files.api";
import type { SequencedProjectRealtimeEvent } from "../../../../api/project-realtime.types";
import { getFileLanguage } from "./fileLanguage";
import { pathReferenceAt, resolveWorkspacePathCandidates } from "./navigationResolver";
import { useFileEditor } from "./useFileEditor";

export function CodeEditor({
  accessToken,
  sandboxId,
  filePath,
  fileName,
  selectedLine,
  selectedColumn,
  onNavigatePath,
  realtimeEvents = [],
  resyncVersion = 0,
}: {
  accessToken: string | null;
  sandboxId: string | null;
  filePath: string | null;
  fileName: string | null;
  selectedLine: number | null;
  selectedColumn: number | null;
  onNavigatePath: (path: string, line: number | null, column: number | null) => void;
  realtimeEvents?: SequencedProjectRealtimeEvent[];
  resyncVersion?: number;
}) {
  const { resolvedTheme } = useTheme();
  const editorState = useFileEditor(sandboxId, filePath, accessToken);
  const saveRef = useRef(editorState.save);
  const refreshRef = useRef(editorState.refreshWithDiff);
  const editorRef = useRef<MonacoEditor.IStandaloneCodeEditor | null>(null);
  const goProviderRef = useRef<{ dispose: () => void } | null>(null);
  const definitionRequestInFlight = useRef(false);
  const processedRealtimeSequence = useRef(0);
  const [showDiff, setShowDiff] = useState(false);
  const [navigationError, setNavigationError] = useState("");
  const navigationContext = useRef({ accessToken, sandboxId, filePath, onNavigatePath });
  navigationContext.current = { accessToken, sandboxId, filePath, onNavigatePath };

  useEffect(() => {
    saveRef.current = editorState.save;
  }, [editorState.save]);

  useEffect(() => {
    refreshRef.current = editorState.refreshWithDiff;
  }, [editorState.refreshWithDiff]);

  useEffect(() => {
    if (!filePath) return;
    const pending = realtimeEvents.filter(({ sequence }) => sequence > processedRealtimeSequence.current);
    const missedEvents = pending.length > 0 && pending[0].sequence > processedRealtimeSequence.current + 1;
    const fileChanged = pending.some(({ event }) =>
      event.path && workspacePath(event.path) === filePath ||
      event.old_path && workspacePath(event.old_path) === filePath,
    );
    if (missedEvents || fileChanged) void refreshRef.current();
    if (pending.length > 0) processedRealtimeSequence.current = pending[pending.length - 1].sequence;
  }, [filePath, realtimeEvents]);

  useEffect(() => {
    if (resyncVersion > 0) void refreshRef.current();
  }, [filePath, resyncVersion]);

  useEffect(() => setShowDiff(false), [filePath]);

  useEffect(() => () => goProviderRef.current?.dispose(), []);

  useEffect(() => {
    if (!selectedLine || !editorRef.current) return;
    editorRef.current.revealLineInCenter(selectedLine);
    editorRef.current.setPosition({ lineNumber: selectedLine, column: selectedColumn ?? 1 });
    editorRef.current.focus();
  }, [filePath, selectedLine, selectedColumn]);

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
        <button
          type="button"
          onClick={() => setShowDiff((current) => !current)}
          disabled={!editorState.diff}
          title={showDiff ? "Show editor" : "Show file diff"}
          aria-label={showDiff ? "Show editor" : "Show file diff"}
          className="grid size-8 place-items-center rounded-md border border-[var(--border)] text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text disabled:cursor-not-allowed disabled:opacity-40"
        >
          <GitCompareArrows size={14} />
        </button>
      </div>
      <div className="min-h-0 flex-1">
        {showDiff && editorState.diff ? (
          <DiffEditor
            width="100%"
            height="100%"
            original={editorState.diff.original}
            modified={editorState.diff.modified}
            language={getFileLanguage(filePath)}
            theme={resolvedTheme === "dark" ? "forge-dark" : "forge-light"}
            options={{ automaticLayout: true, readOnly: true, minimap: { enabled: false }, scrollBeyondLastLine: false }}
          />
        ) : <Editor
          width="100%"
          height="100%"
          path={filePath}
          language={getFileLanguage(filePath)}
          value={editorState.content}
          onChange={(value) => editorState.updateDraft(value ?? "")}
          onMount={(editor, monaco) => {
            editorRef.current = editor;
            goProviderRef.current?.dispose();
            goProviderRef.current = monaco.languages.registerDefinitionProvider("go", {
              provideDefinition: async (
                _model: Monaco.editor.ITextModel,
                position: Monaco.Position,
                token: Monaco.CancellationToken,
              ) => {
                const current = navigationContext.current;
                if (!current.accessToken || !current.sandboxId || !current.filePath) return null;
                if (definitionRequestInFlight.current) return null;
                definitionRequestInFlight.current = true;
                try {
                  const definition = await getGoDefinition(
                    current.accessToken,
                    current.sandboxId,
                    current.filePath,
                    position.lineNumber,
                    position.column,
                  );
                  if (token.isCancellationRequested) return null;
                  current.onNavigatePath(definition.path, definition.line, definition.column);
                  return {
                    uri: monaco.Uri.file(definition.path),
                    range: new monaco.Range(definition.line, definition.column, definition.line, definition.column),
                  };
                } catch {
                  if (!token.isCancellationRequested) setNavigationError("Go definition could not be resolved.");
                  return null;
                } finally {
                  definitionRequestInFlight.current = false;
                }
              },
            });
            if (selectedLine) {
              editor.revealLineInCenter(selectedLine);
              editor.setPosition({ lineNumber: selectedLine, column: selectedColumn ?? 1 });
            }
            async function openPathReference(position: { lineNumber: number; column: number } | null) {
              const current = navigationContext.current;
              const model = editor.getModel();
              if (!position || !model || !current.filePath || !current.accessToken || !current.sandboxId) return false;
              const reference = pathReferenceAt(model.getLineContent(position.lineNumber), position.column);
              if (!reference) return false;
              for (const candidate of resolveWorkspacePathCandidates(current.filePath, reference)) {
                try {
                  await readFile(current.accessToken, current.sandboxId, candidate);
                  current.onNavigatePath(candidate, reference.line, reference.column);
                  setNavigationError("");
                  return true;
                } catch {
                  continue;
                }
              }
              setNavigationError(`Workspace path not found: ${reference.path}`);
              return true;
            }

            editor.onMouseDown((event) => {
              const browserEvent = event.event.browserEvent;
              if (!browserEvent.ctrlKey && !browserEvent.metaKey) return;
              const position = event.target.position;
              if (!position) return;
              const model = editor.getModel();
              if (!model || !pathReferenceAt(model.getLineContent(position.lineNumber), position.column)) return;
              browserEvent.preventDefault();
              void openPathReference(position);
            });
            editor.addAction({
              id: "forge.go-to-file-or-definition",
              label: "Go to File or Definition",
              keybindings: [monaco.KeyCode.F12, monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter],
              run: async () => {
                if (!await openPathReference(editor.getPosition())) {
                  await editor.getAction("editor.action.revealDefinition")?.run();
                }
              },
            });
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
        />}
      </div>
      {navigationError && (
        <div role="status" className="border-t border-[var(--border)] px-4 py-2 text-xs text-forge-signal">
          {navigationError}
        </div>
      )}
      {editorState.externalChange && (
        <div className="flex items-center justify-between gap-3 border-t border-[var(--border)] px-4 py-2 text-xs text-forge-signal">
          <span>The file changed in the workspace; your unsaved draft is preserved.</span>
          <button type="button" className="shrink-0 underline underline-offset-2" onClick={editorState.useExternalVersion}>
            Load workspace version
          </button>
        </div>
      )}
    </div>
  );
}

function workspacePath(path: string) {
  const normalized = path.replaceAll("\\", "/").replace(/^\/+/, "");
  return normalized.startsWith("workspace/") ? `/${normalized}` : `/workspace/${normalized}`;
}