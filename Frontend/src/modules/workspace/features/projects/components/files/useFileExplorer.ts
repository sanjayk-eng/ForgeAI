import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useRef, useState } from "react";
import { listFiles, type FileEntry } from "../../../../api/sandbox.api";
import {
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
  };
}