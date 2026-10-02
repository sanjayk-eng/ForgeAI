import { useEffect, useRef, useState } from "react";
import { Tree } from "react-arborist";
import { FilePlus2, FolderPlus } from "lucide-react";
import type { FileEntry } from "../../../../api/sandbox-files.api";
import type { SequencedProjectRealtimeEvent } from "../../../../api/project-realtime.types";
import { FileActionDialog, type FileActionDialogState } from "./FileActionDialog";
import { FileExplorerActionsContext } from "./FileExplorerActionsContext";
import { FileExplorerNode } from "./FileExplorerNode";
import {
  getFileName,
  findFileTreeItem,
  getParentPath,
  joinFilePath,
  WORKSPACE_ROOT,
  type FileTreeItem,
} from "./fileTree";
import { useFileExplorer } from "./useFileExplorer";

interface FileExplorerProps {
  accessToken: string;
  sandboxId: string;
  selectedPath?: string | null;
  onSelect: (file: FileEntry) => void;
  onPathChanged?: (oldPath: string, newPath: string) => void;
  onPathDeleted?: (path: string) => void;
  realtimeEvents?: SequencedProjectRealtimeEvent[];
  resyncVersion?: number;
  changedPaths?: string[];
}

export function FileExplorer({
  accessToken,
  sandboxId,
  selectedPath,
  onSelect,
  onPathChanged,
  onPathDeleted,
  realtimeEvents = [],
  resyncVersion = 0,
  changedPaths = [],
}: FileExplorerProps) {
  const treeContainer = useRef<HTMLDivElement>(null);
  const [treeHeight, setTreeHeight] = useState(1);
  const [dialogAction, setDialogAction] = useState<FileActionDialogState | null>(null);
  const [dialogError, setDialogError] = useState("");
  const [pendingCreate, setPendingCreate] = useState<{
    parentPath: string;
    entryType: "file" | "directory";
  } | null>(null);
  const [createName, setCreateName] = useState("");
  const [createError, setCreateError] = useState("");
  const explorer = useFileExplorer(sandboxId, accessToken);
  const processedRealtimeSequence = useRef(0);

  useEffect(() => {
    const pending = realtimeEvents.filter(({ sequence }) => sequence > processedRealtimeSequence.current);
    if (pending.length === 0) return;
    if (pending[0].sequence > processedRealtimeSequence.current + 1) {
      void explorer.refreshLoadedDirectories();
    }
    for (const update of pending) {
      explorer.applyRealtimeEvent(update.event);
      processedRealtimeSequence.current = update.sequence;
    }
  }, [explorer.applyRealtimeEvent, explorer.refreshLoadedDirectories, realtimeEvents]);

  useEffect(() => {
    if (resyncVersion > 0) void explorer.refreshLoadedDirectories();
  }, [explorer.refreshLoadedDirectories, resyncVersion]);

  useEffect(() => {
    const element = treeContainer.current;
    if (!element) return;
    const observer = new ResizeObserver(([entry]) => {
      setTreeHeight(Math.max(1, Math.floor(entry.contentRect.height)));
    });
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  function openDialog(action: FileActionDialogState) {
    setDialogError("");
    setDialogAction(action);
  }

  function startCreate(parentPath: string, entryType: "file" | "directory") {
    setCreateName("");
    setCreateError("");
    setPendingCreate({ parentPath, entryType });
  }

  async function submitCreate() {
    if (!pendingCreate || !createName.trim()) return;
    setCreateError("");
    try {
      await explorer.createEntry(pendingCreate.parentPath, pendingCreate.entryType, createName.trim());
      setPendingCreate(null);
    } catch (error) {
      setCreateError(error instanceof Error ? error.message : "Could not create item");
    }
  }

  const actions = {
    createEntry: startCreate,
    createName,
    createError,
    isCreating: explorer.isPending,
    setCreateName,
    submitCreate,
    cancelCreate: () => setPendingCreate(null),
    deleteEntry: (file: FileTreeItem) => openDialog({ kind: "delete", file }),
    renameEntry: (file: FileTreeItem) => openDialog({ kind: "rename", file }),
    selectFile: onSelect,
  };

  async function confirmDialog(name?: string) {
    if (!dialogAction) return;
    setDialogError("");
    try {
      if (dialogAction.kind === "rename") {
        if (!name?.trim()) return;
        const { file } = dialogAction;
        await explorer.renameEntry(file.path, name.trim());
        onPathChanged?.(file.path, joinFilePath(getParentPath(file.path), name.trim()));
      } else {
        await explorer.deleteEntry(dialogAction.file.path);
        onPathDeleted?.(dialogAction.file.path);
      }
      setDialogAction(null);
    } catch (error) {
      setDialogError(error instanceof Error ? error.message : "File operation failed");
    }
  }

  function addCreatePlaceholder(entries: FileTreeItem[]): FileTreeItem[] {
    if (!pendingCreate) return entries;
    const placeholder: FileTreeItem = {
      name: "",
      path: `__forgeai_create__:${pendingCreate.parentPath}`,
      is_directory: false,
      isCreatePlaceholder: true,
      createEntryType: pendingCreate.entryType,
    };
    if (pendingCreate.parentPath === WORKSPACE_ROOT) {
      return [placeholder, ...entries];
    }
    return entries.map((entry) => {
      if (entry.path === pendingCreate.parentPath && entry.is_directory) {
        return { ...entry, children: [placeholder, ...(entry.children ?? [])] };
      }
      if (entry.is_directory && entry.children) {
        return { ...entry, children: addCreatePlaceholder(entry.children) };
      }
      return entry;
    });
  }

  const changedPathSet = new Set(changedPaths);
  const markChanged = (entries: FileTreeItem[]): FileTreeItem[] => entries.map((entry) => ({
    ...entry,
    is_modified: !entry.is_directory && changedPathSet.has(entry.path),
    ...(entry.children ? { children: markChanged(entry.children) } : {}),
  }));
  const treeData = addCreatePlaceholder(markChanged(explorer.data));

  return (
    <FileExplorerActionsContext.Provider value={actions}>
      <div className="flex h-full min-h-0 flex-col gap-2">
        <div className="flex items-center justify-end gap-1">
          <button
            type="button"
            title="Create file"
            aria-label="Create file in workspace"
            className="grid size-7 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"
            onClick={() => startCreate(WORKSPACE_ROOT, "file")}
          >
            <FilePlus2 size={14} />
          </button>
          <button
            type="button"
            title="Create folder"
            aria-label="Create folder in workspace"
            className="grid size-7 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"
            onClick={() => startCreate(WORKSPACE_ROOT, "directory")}
          >
            <FolderPlus size={14} />
          </button>
        </div>
        {explorer.isLoading && (
          <div className="px-2 py-1 text-xs text-forge-muted">Loading files...</div>
        )}
        {!explorer.isLoading && explorer.rootError && (
          <div className="px-2 py-1 text-xs text-forge-signal">Unable to load files</div>
        )}
        {!explorer.isLoading && !explorer.rootError && treeData.length === 0 && (
          <div className="px-2 py-1 text-xs text-forge-muted">No files found</div>
        )}
        {explorer.operationError && (
          <div className="px-2 py-1 text-xs text-forge-signal">A file operation failed</div>
        )}
        <div ref={treeContainer} className="min-h-0 flex-1">
          {treeData.length > 0 && (
            <Tree<FileTreeItem>
              data={treeData}
              idAccessor="path"
              selection={selectedPath ?? undefined}
              width="100%"
              height={treeHeight}
              rowHeight={28}
              indent={14}
              overscanCount={6}
              disableMultiSelection
              onToggle={(path) => void explorer.loadChildren(path)}
              onDelete={({ ids }) => {
                const file = findFileTreeItem(explorer.data, ids[0]);
                if (file) actions.deleteEntry(file);
              }}
              onMove={async ({ dragIds, parentId }) => {
                const parentPath = parentId ?? WORKSPACE_ROOT;
                const moves = dragIds
                  .map((oldPath) => ({
                    oldPath,
                    newPath: joinFilePath(parentPath, getFileName(oldPath)),
                  }))
                  .filter(({ oldPath, newPath }) => oldPath !== newPath);
                if (!moves.length) return;
                await explorer.moveEntries(moves);
                moves.forEach(({ oldPath, newPath }) =>
                  onPathChanged?.(oldPath, newPath),
                );
              }}
              disableDrop={({ parentNode }) =>
                !parentNode.isRoot && !parentNode.data.is_directory
              }
              aria-label="Project files"
            >
              {FileExplorerNode}
            </Tree>
          )}
        </div>
      </div>
      {dialogAction && (
        <FileActionDialog
          key={`${dialogAction.kind}:${dialogAction.file.path}`}
          action={dialogAction}
          error={dialogError}
          isPending={explorer.isPending}
          onClose={() => setDialogAction(null)}
          onConfirm={confirmDialog}
        />
      )}
    </FileExplorerActionsContext.Provider>
  );
}