import { useEffect, useRef, useState, type FormEvent } from "react";
import { Tree } from "react-arborist";
import { Check, FilePlus2, FolderPlus, X } from "lucide-react";
import type { FileEntry } from "../../../../api/sandbox.api";
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
}

export function FileExplorer({
  accessToken,
  sandboxId,
  selectedPath,
  onSelect,
  onPathChanged,
  onPathDeleted,
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

  async function submitCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
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
        {!explorer.isLoading && !explorer.rootError && explorer.data.length === 0 && (
          <div className="px-2 py-1 text-xs text-forge-muted">No files found</div>
        )}
        {pendingCreate && (
          <form
            className="grid gap-1 rounded-md border border-forge-accent/40 bg-forge-accent/[0.06] p-2"
            onSubmit={submitCreate}
          >
            <div className="flex min-w-0 items-center gap-2">
              {pendingCreate.entryType === "file"
                ? <FilePlus2 size={14} className="shrink-0 text-forge-accent" />
                : <FolderPlus size={14} className="shrink-0 text-forge-accent" />}
              <input
                autoFocus
                aria-label={pendingCreate.entryType === "file" ? "New file name" : "New folder name"}
                className="h-8 min-w-0 flex-1 border-0 bg-transparent text-xs text-forge-text outline-none placeholder:text-forge-muted focus:ring-0"
                placeholder={pendingCreate.entryType === "file" ? "File name, e.g. notes.md" : "Folder name"}
                value={createName}
                onChange={(event) => setCreateName(event.target.value)}
                disabled={explorer.isPending}
              />
              <button
                type="submit"
                title="Create"
                aria-label="Create"
                className="grid size-7 shrink-0 place-items-center rounded text-forge-accent hover:bg-forge-accent/10 disabled:opacity-50"
                disabled={!createName.trim() || explorer.isPending}
              >
                <Check size={15} />
              </button>
              <button
                type="button"
                title="Cancel"
                aria-label="Cancel"
                className="grid size-7 shrink-0 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)]"
                onClick={() => setPendingCreate(null)}
                disabled={explorer.isPending}
              >
                <X size={15} />
              </button>
            </div>
            <span className="truncate pl-6 text-[10px] text-forge-muted">
              {pendingCreate.entryType === "file" ? "File" : "Folder"} in {pendingCreate.parentPath}
            </span>
            {createError && <span className="pl-6 text-xs text-forge-signal" role="alert">{createError}</span>}
          </form>
        )}
        {explorer.operationError && (
          <div className="px-2 py-1 text-xs text-forge-signal">A file operation failed</div>
        )}
        <div ref={treeContainer} className="min-h-0 flex-1">
          {explorer.data.length > 0 && (
            <Tree<FileTreeItem>
              data={explorer.data}
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