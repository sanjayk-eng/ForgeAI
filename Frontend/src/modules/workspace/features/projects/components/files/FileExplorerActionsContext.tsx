import { createContext, useContext } from "react";
import type { FileTreeItem } from "./fileTree";

interface FileExplorerActions {
  createEntry: (parentPath: string, type: "file" | "directory") => void;
  deleteEntry: (file: FileTreeItem) => void;
  renameEntry: (file: FileTreeItem) => void;
  selectFile: (file: FileTreeItem) => void;
}

export const FileExplorerActionsContext = createContext<FileExplorerActions | null>(null);

export function useFileExplorerActions() {
  const actions = useContext(FileExplorerActionsContext);
  if (!actions) throw new Error("File explorer actions are unavailable");
  return actions;
}