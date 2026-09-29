import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createFileOrDirectory,
  deletePath,
  renamePath,
} from "../../../../api/sandbox-files.api";
import { getParentPath, joinFilePath, sandboxFileKeys } from "./fileTree";

interface FileMove {
  oldPath: string;
  newPath: string;
}

export function useFileOperations(
  sandboxId: string,
  accessToken: string,
  onChanged: () => void,
) {
  const queryClient = useQueryClient();

  function invalidateFiles() {
    onChanged();
    void queryClient.invalidateQueries({
      queryKey: sandboxFileKeys.allFiles(sandboxId),
    });
    void queryClient.invalidateQueries({
      queryKey: sandboxFileKeys.allContent(sandboxId),
    });
  }

  const createMutation = useMutation({
    mutationFn: ({
      parentPath,
      name,
      type,
    }: {
      parentPath: string;
      name: string;
      type: "file" | "directory";
    }) =>
      createFileOrDirectory(
        accessToken,
        sandboxId,
        joinFilePath(parentPath, name.trim()),
        type,
      ),
    onSuccess: invalidateFiles,
  });

  const renameMutation = useMutation({
    mutationFn: ({ path, name }: { path: string; name: string }) =>
      renamePath(
        accessToken,
        sandboxId,
        path,
        joinFilePath(getParentPath(path), name.trim()),
      ),
    onSuccess: invalidateFiles,
  });

  const deleteMutation = useMutation({
    mutationFn: (path: string) => deletePath(accessToken, sandboxId, path),
    onSuccess: invalidateFiles,
  });

  const moveMutation = useMutation({
    mutationFn: async (moves: FileMove[]) => {
      for (const move of moves) {
        await renamePath(accessToken, sandboxId, move.oldPath, move.newPath);
      }
    },
    onSuccess: invalidateFiles,
    onError: invalidateFiles,
  });

  return {
    createEntry: (parentPath: string, type: "file" | "directory", name: string) =>
      createMutation.mutateAsync({ parentPath, type, name }),
    renameEntry: (path: string, name: string) =>
      renameMutation.mutateAsync({ path, name }),
    deleteEntry: (path: string) => deleteMutation.mutateAsync(path),
    moveEntries: (moves: FileMove[]) => moveMutation.mutateAsync(moves),
    error:
      createMutation.error ??
      renameMutation.error ??
      deleteMutation.error ??
      moveMutation.error,
    isPending:
      createMutation.isPending ||
      renameMutation.isPending ||
      deleteMutation.isPending ||
      moveMutation.isPending,
  };
}