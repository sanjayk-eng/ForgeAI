import {
  AlertCircle,
  Check,
  ChevronDown,
  ChevronRight,
  FileCode2,
  FilePlus2,
  Folder,
  FolderOpen,
  FolderPlus,
  LoaderCircle,
  Pencil,
  Trash2,
  X,
} from "lucide-react";
import type { NodeRendererProps } from "react-arborist";
import type { FileTreeItem } from "./fileTree";
import { useFileExplorerActions } from "./FileExplorerActionsContext";

export function FileExplorerNode({
  node,
  style,
  dragHandle,
}: NodeRendererProps<FileTreeItem>) {
  const actions = useFileExplorerActions();
  const file = node.data;
  const isDirectory = file.is_directory;

  if (file.isCreatePlaceholder) {
    const entryType = file.createEntryType ?? "file";
    return (
      <div
        style={style}
        className="group flex min-w-0 items-center gap-1.5 rounded-md border border-[var(--border)] bg-[var(--surface)] px-1.5 py-1 text-xs shadow-[inset_0_1px_0_rgba(255,255,255,0.04)]"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="flex shrink-0 items-center gap-1 text-forge-accent">
          {entryType === "file" ? <FilePlus2 size={13} /> : <FolderPlus size={13} />}
        </div>

        <form
          className="flex min-w-0 flex-1 items-center gap-1.5"
          onSubmit={(event) => {
            event.preventDefault();
            event.stopPropagation();
            void actions.submitCreate();
          }}
        >
          <span className="flex shrink-0 items-center">
            {entryType === "file" ? <FileCode2 size={13} className="text-forge-muted" /> : <FolderOpen size={13} className="text-forge-accent" />}
          </span>
          <input
            autoFocus
            aria-label={entryType === "file" ? "New file name" : "New folder name"}
            className="h-7 min-w-0 flex-1 border-0 bg-transparent text-xs font-medium text-forge-text outline-none placeholder:text-forge-muted focus:ring-0"
            placeholder={entryType === "file" ? "File name" : "Folder name"}
            value={actions.createName}
            onChange={(event) => actions.setCreateName(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Escape") actions.cancelCreate();
            }}
            disabled={actions.isCreating}
          />
        </form>

        <div className="flex shrink-0 items-center gap-1">
          <button
            type="submit"
            title="Create"
            aria-label="Create"
            className="grid size-6 shrink-0 place-items-center rounded-md bg-forge-accent/10 text-forge-accent hover:bg-forge-accent/15 disabled:cursor-not-allowed disabled:opacity-50"
            disabled={!actions.createName.trim() || actions.isCreating}
            onClick={() => void actions.submitCreate()}
          >
            <Check size={13} />
          </button>
          <button
            type="button"
            title="Cancel"
            aria-label="Cancel"
            className="grid size-6 shrink-0 place-items-center rounded-md text-forge-muted hover:bg-[var(--surface-hover)] disabled:cursor-not-allowed disabled:opacity-50"
            onClick={actions.cancelCreate}
            disabled={actions.isCreating}
          >
            <X size={13} />
          </button>
        </div>

        {actions.createError && (
          <span title={actions.createError} className="flex shrink-0 items-center">
            <AlertCircle size={13} className="text-forge-signal" />
          </span>
        )}
      </div>
    );
  }

  return (
    <div
      ref={dragHandle}
      style={style}
      className={`group flex min-w-0 items-center gap-1 rounded px-1 text-xs ${node.isSelected ? "bg-forge-accent/15 text-forge-text" : "text-forge-soft hover:bg-[var(--surface-hover)] hover:text-forge-text"}`}
    >
      <button
        type="button"
        aria-label={isDirectory ? `${node.isOpen ? "Collapse" : "Expand"} ${file.name}` : `Open ${file.name}`}
        className="flex min-w-0 flex-1 items-center gap-1.5 overflow-hidden text-left outline-none"
        onClick={(event) => {
          node.handleClick(event);
          if (isDirectory) node.toggle();
          else actions.selectFile(file);
        }}
      >
        {isDirectory ? (
          node.isOpen ? <ChevronDown size={13} className="shrink-0 text-forge-muted" /> : <ChevronRight size={13} className="shrink-0 text-forge-muted" />
        ) : (
          <span className="w-[13px] shrink-0" />
        )}
        {isDirectory ? (
          node.isOpen ? <FolderOpen size={14} className="shrink-0 text-forge-accent" /> : <Folder size={14} className="shrink-0 text-forge-muted" />
        ) : (
          <FileCode2 size={14} className="shrink-0 text-forge-muted" />
        )}
        <span className="truncate">{file.name}</span>
      </button>

      {file.childrenLoading && <LoaderCircle size={12} className="animate-spin text-forge-muted" />}
      {file.childrenError && <AlertCircle size={12} className="text-forge-signal" aria-label="Unable to load folder" />}

      <div className="flex shrink-0 items-center opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100">
        {isDirectory && (
          <>
            <button
              type="button"
              title="Create file in folder"
              aria-label={`Create file in ${file.name}`}
              className="grid size-6 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"
              onClick={(event) => {
                event.stopPropagation();
                node.open();
                actions.createEntry(file.path, "file");
              }}
            >
              <FilePlus2 size={12} />
            </button>
            <button
              type="button"
              title="Create folder"
              aria-label={`Create folder in ${file.name}`}
              className="grid size-6 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"
              onClick={(event) => {
                event.stopPropagation();
                node.open();
                actions.createEntry(file.path, "directory");
              }}
            >
              <FolderPlus size={12} />
            </button>
          </>
        )}
        <button
          type="button"
          title="Rename"
          aria-label={`Rename ${file.name}`}
          className="grid size-6 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"
          onClick={(event) => {
            event.stopPropagation();
            actions.renameEntry(file);
          }}
        >
          <Pencil size={12} />
        </button>
        <button
          type="button"
          title="Delete"
          aria-label={`Delete ${file.name}`}
          className="grid size-6 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-signal"
          onClick={(event) => {
            event.stopPropagation();
            actions.deleteEntry(file);
          }}
        >
          <Trash2 size={12} />
        </button>
      </div>
    </div>
  );
}