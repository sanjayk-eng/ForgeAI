import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useRef, useState } from "react";
import { listFiles, type FileEntry } from "../../../../api/sandbox-files.api";
import type { ProjectRealtimeEvent } from "../../../../api/project-realtime.types";
import {
  findFileTreeItem,
  getParentPath,
  isWorkspaceEntryVisible,
  sandboxFileKeys,
  WORKSPACE_ROOT,
  type FileTreeItem,
} from "./fileTree";
import { useFileOperations } from "./useFileOperations";

export function useFileExplorer(
  sandboxId: string,
  accessToken: string,
) {
  const queryClient = useQueryClient();
  const loadedPaths = useRef(new Set<string>());
  const loadingPaths = useRef(new Set<string>());
  const [childrenByPath, setChildrenByPath] = useState<Record<string, FileEntry[]>>({});
  const [loadingByPath, setLoadingByPath] = useState<Record<string, boolean>>({});
  const [errorsByPath, setErrorsByPath] = useState<Record<string, boolean>>({});

  const rootQuery = useQuery({
    queryKey: sandboxFileKeys.directory(sandboxId, WORKSPACE_ROOT),
    queryFn: () => listFiles(accessToken, sandboxId, WORKSPACE_ROOT),
    enabled: Boolean(sandboxId && accessToken),
    staleTime: 30_000,
  });

  const resetChildren = useCallback(() => {
    loadedPaths.current.clear();
    setChildrenByPath({});
    setLoadingByPath({});
    setErrorsByPath({});
  }, []);

  const operations = useFileOperations(sandboxId, accessToken, resetChildren);

  const loadChildren = useCallback(async (path: string) => {
    if (loadedPaths.current.has(path) || loadingPaths.current.has(path)) return;

    loadingPaths.current.add(path);
    setLoadingByPath((current) => ({ ...current, [path]: true }));
    setErrorsByPath((current) => ({ ...current, [path]: false }));
    try {
      const result = await queryClient.fetchQuery({
        queryKey: sandboxFileKeys.directory(sandboxId, path),
        queryFn: () => listFiles(accessToken, sandboxId, path),
        staleTime: 30_000,
      });
      loadedPaths.current.add(path);
      setChildrenByPath((current) => ({ ...current, [path]: result.files }));
    } catch {
      setErrorsByPath((current) => ({ ...current, [path]: true }));
    } finally {
      loadingPaths.current.delete(path);
      setLoadingByPath((current) => ({ ...current, [path]: false }));
    }
  }, [accessToken, queryClient, sandboxId]);

  const refreshLoadedDirectories = useCallback(async () => {
    await queryClient.fetchQuery({
      queryKey: sandboxFileKeys.directory(sandboxId, WORKSPACE_ROOT),
      queryFn: () => listFiles(accessToken, sandboxId, WORKSPACE_ROOT),
      staleTime: 0,
    });
    const paths = [...loadedPaths.current];
    await Promise.all(paths.map(async (path) => {
      try {
        const result = await queryClient.fetchQuery({
          queryKey: sandboxFileKeys.directory(sandboxId, path),
          queryFn: () => listFiles(accessToken, sandboxId, path),
          staleTime: 0,
        });
        setChildrenByPath((current) => ({ ...current, [path]: result.files }));
      } catch {
        setErrorsByPath((current) => ({ ...current, [path]: true }));
      }
    }));
  }, [accessToken, queryClient, sandboxId]);

  const applyRealtimeEvent = useCallback((event: ProjectRealtimeEvent) => {
    if (!event.path || !event.event.startsWith("file.")) return;
    const path = workspaceAbsolutePath(event.path);
    const oldPath = event.old_path ? workspaceAbsolutePath(event.old_path) : "";

    const updateDirectory = (directoryPath: string, update: (entries: FileEntry[]) => FileEntry[]) => {
      const queryKey = sandboxFileKeys.directory(sandboxId, directoryPath);
      queryClient.setQueryData<{ files: FileEntry[] }>(queryKey, (current) =>
        current ? { ...current, files: update(current.files) } : current,
      );
      if (directoryPath !== WORKSPACE_ROOT) {
        setChildrenByPath((current) => current[directoryPath]
          ? { ...current, [directoryPath]: update(current[directoryPath]) }
          : current,
        );
      }
    };

    const removeEntry = (entryPath: string, keepLoadedChildren = false) => {
      updateDirectory(getParentPath(entryPath), (entries) => entries.filter((entry) => entry.path !== entryPath));
      loadedPaths.current.delete(entryPath);
      if (!keepLoadedChildren) {
        setChildrenByPath((current) => Object.fromEntries(
          Object.entries(current).filter(([directory]) => directory !== entryPath && !directory.startsWith(`${entryPath}/`)),
        ));
        void queryClient.removeQueries({ queryKey: sandboxFileKeys.directory(sandboxId, entryPath) });
      }
    };

    if (event.event === "file.deleted") {
      removeEntry(path);
      return;
    }

    if (event.event === "file.renamed" && oldPath) {
      const existing = findFileTreeItem(buildTree(rootQuery.data?.files ?? []), oldPath);
      const childrenToMove = Object.entries(childrenByPath)
        .filter(([directory]) => directory === oldPath || directory.startsWith(`${oldPath}/`));
      removeEntry(oldPath, true);
      insertEntry(path, event.is_directory ?? existing?.is_directory ?? false, existing, updateDirectory);
      if (childrenToMove.length > 0) {
        const moved = childrenToMove.map(([directory, entries]) => {
          const newDirectory = `${path}${directory.slice(oldPath.length)}`;
          const newEntries = entries.map((entry) => ({
            ...entry,
            path: entry.path === oldPath || entry.path.startsWith(`${oldPath}/`)
              ? `${path}${entry.path.slice(oldPath.length)}`
              : entry.path,
          }));
          queryClient.setQueryData(sandboxFileKeys.directory(sandboxId, newDirectory), { files: newEntries });
          return [newDirectory, newEntries] as const;
        });
        setChildrenByPath((current) => {
          const next = { ...current };
          for (const [directory] of childrenToMove) delete next[directory];
          for (const [directory, entries] of moved) next[directory] = entries;
          return next;
        });
        for (const [directory] of childrenToMove) {
          loadedPaths.current.delete(directory);
          void queryClient.removeQueries({ queryKey: sandboxFileKeys.directory(sandboxId, directory) });
        }
        for (const [directory] of moved) loadedPaths.current.add(directory);
      }
      return;
    }

    if (event.event === "file.created") {
      insertEntry(path, event.is_directory ?? false, undefined, updateDirectory);
    }
  }, [childrenByPath, queryClient, rootQuery.data?.files, sandboxId]);

  function buildTree(entries: FileEntry[]): FileTreeItem[] {
    return entries.filter(isWorkspaceEntryVisible).map((entry) => ({
      ...entry,
      ...(entry.is_directory
        ? {
            children: buildTree(childrenByPath[entry.path] ?? []),
            childrenLoading: loadingByPath[entry.path] ?? false,
            childrenError: errorsByPath[entry.path] ?? false,
          }
        : {}),
    }));
  }

  return {
    ...operations,
    data: buildTree(rootQuery.data?.files ?? []),
    rootError: rootQuery.error,
    operationError: operations.error,
    isLoading: rootQuery.isLoading,
    loadChildren,
    applyRealtimeEvent,
    refreshLoadedDirectories,
  };
}

function workspaceAbsolutePath(path: string) {
  const normalized = path.replaceAll("\\", "/").replace(/^\/+/, "");
  return normalized === WORKSPACE_ROOT.slice(1) || normalized.startsWith(`${WORKSPACE_ROOT.slice(1)}/`)
    ? `/${normalized}`
    : `${WORKSPACE_ROOT}/${normalized}`;
}

function insertEntry(
  path: string,
  isDirectory: boolean,
  existing: FileEntry | undefined,
  updateDirectory: (directoryPath: string, update: (entries: FileEntry[]) => FileEntry[]) => void,
) {
  const entry: FileEntry = {
    name: path.split("/").at(-1) ?? path,
    path,
    is_directory: isDirectory,
    ...(existing?.size === undefined ? {} : { size: existing.size }),
    ...(existing?.modified_at === undefined ? {} : { modified_at: existing.modified_at }),
  };
  updateDirectory(getParentPath(path), (entries) => {
    const next = [...entries.filter((item) => item.path !== path), entry];
    return next.sort((left, right) => left.name.localeCompare(right.name));
  });
}