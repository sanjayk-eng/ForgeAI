import { useEffect, useState, type FormEvent } from "react";
import { Pencil, Trash2, X } from "lucide-react";
import type { FileTreeItem } from "./fileTree";

export type FileActionDialogState =
  | { kind: "rename"; file: FileTreeItem }
  | { kind: "delete"; file: FileTreeItem };

export function FileActionDialog({
  action,
  error,
  isPending,
  onClose,
  onConfirm,
}: {
  action: FileActionDialogState;
  error: string;
  isPending: boolean;
  onClose: () => void;
  onConfirm: (name?: string) => void;
}) {
  const initialName = action.kind === "rename" ? action.file.name : "";
  const [name, setName] = useState(initialName);
  const isDelete = action.kind === "delete";
  const title = isDelete ? "Delete item" : "Rename item";

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape" && !isPending) onClose();
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isPending, onClose]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isPending) return;
    if (!isDelete && !name.trim()) return;
    onConfirm(isDelete ? undefined : name.trim());
  }

  return (
    <div
      className="fixed inset-0 z-50 grid place-items-center bg-[var(--overlay)] p-4 backdrop-blur-sm"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget && !isPending) onClose();
      }}
    >
      <form
        aria-labelledby="file-action-title"
        aria-modal="true"
        className="w-full max-w-[420px] rounded-lg border border-[var(--border)] bg-forge-panel p-5 shadow-[0_24px_80px_rgba(0,0,0,.3)]"
        onSubmit={submit}
        role="dialog"
      >
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-center gap-2 text-sm font-semibold text-forge-text">
            {action.kind === "rename" && <Pencil size={16} className="text-forge-accent" />}
            {action.kind === "delete" && <Trash2 size={16} className="text-forge-signal" />}
            <h2 id="file-action-title" className="m-0 text-sm font-semibold">{title}</h2>
          </div>
          <button
            type="button"
            aria-label="Close dialog"
            className="grid size-7 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"
            disabled={isPending}
            onClick={onClose}
          >
            <X size={15} />
          </button>
        </div>

        {isDelete ? (
          <p className="mt-4 text-sm text-forge-soft">
            Delete <span className="font-semibold text-forge-text">{action.file.name}</span>? This cannot be undone.
          </p>
        ) : (
          <label className="mt-5 block text-xs font-medium text-forge-soft">
            <span>Name</span>
            <input
              autoFocus
              className="input mt-2"
              maxLength={255}
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="New name"
              disabled={isPending}
            />
          </label>
        )}

        {error && <p className="mt-3 text-xs text-forge-signal" role="alert">{error}</p>}

        <div className="mt-6 flex justify-end gap-2">
          <button
            type="button"
            className="rounded-md border border-[var(--border)] px-3 py-2 text-xs font-semibold text-forge-soft hover:bg-[var(--surface-hover)]"
            disabled={isPending}
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            type="submit"
            className={`rounded-md px-3 py-2 text-xs font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50 ${isDelete ? "bg-forge-signal hover:brightness-110" : "bg-forge-accent hover:bg-forge-accent-strong"}`}
            disabled={isPending || (!isDelete && !name.trim())}
          >
            {isPending ? "Working..." : isDelete ? "Delete" : "Rename"}
          </button>
        </div>
      </form>
    </div>
  );
}