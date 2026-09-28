import type { FileEntry } from "../../../../api/sandbox.api";

export const WORKSPACE_ROOT = "/workspace";

export interface FileTreeItem extends FileEntry {
  children?: FileTreeItem[];
  childrenLoading?: boolean;
  childrenError?: boolean;
}

export const sandboxFileKeys = {
  allFiles: (sandboxId: string) => ["sandbox-files", sandboxId] as const,
  directory: (sandboxId: string, path: string) =>
    ["sandbox-files", sandboxId, path] as const,
  allContent: (sandboxId: string) => ["sandbox-file-content", sandboxId] as const,
  content: (sandboxId: string, path: string) =>
    ["sandbox-file-content", sandboxId, path] as const,
};

export function joinFilePath(parentPath: string, name: string) {
  return `${parentPath.replace(/\/$/, "")}/${name}`;
}

export function getFileName(path: string) {
  return path.split("/").filter(Boolean).at(-1) ?? path;
}

export function getParentPath(path: string) {
  const segments = path.split("/").filter(Boolean);
  segments.pop();
  return segments.length ? `/${segments.join("/")}` : "/";
}

const hiddenWorkspaceEntries = new Set([".git", ".forgeai-repository-cloned"]);

export function isWorkspaceEntryVisible(entry: FileEntry) {
  return !hiddenWorkspaceEntries.has(entry.name);
}

export function findFileTreeItem(entries: FileTreeItem[], path: string): FileTreeItem | undefined {
  for (const entry of entries) {
    if (entry.path === path) return entry;
    const child = entry.children && findFileTreeItem(entry.children, path);
    if (child) return child;
  }
  return undefined;
}