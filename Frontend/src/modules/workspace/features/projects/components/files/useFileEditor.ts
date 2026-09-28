import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { readFile, saveFile } from "../../../../api/sandbox.api";
import { sandboxFileKeys } from "./fileTree";

interface FileDraft {
  path: string;
  content: string;
}

export function useFileEditor(
  sandboxId: string | null,
  filePath: string | null,
  accessToken: string | null,
) {
  const queryClient = useQueryClient();
  const [draft, setDraft] = useState<FileDraft | null>(null);

  const fileQuery = useQuery({
    queryKey: sandboxId && filePath
      ? sandboxFileKeys.content(sandboxId, filePath)
      : ["sandbox-file-content", sandboxId, filePath],
    queryFn: () => {
      if (!sandboxId || !filePath || !accessToken) {
        throw new Error("Missing sandbox, path or access token");
      }
      return readFile(accessToken, sandboxId, filePath);
    },
    enabled: Boolean(sandboxId && filePath && accessToken),
    staleTime: 30_000,
  });

  const content = filePath && draft?.path === filePath
    ? draft.content
    : fileQuery.data?.content ?? "";
  const hasUnsavedChanges = Boolean(
    filePath &&
    draft?.path === filePath &&
    fileQuery.data &&
    draft.content !== fileQuery.data.content,
  );

  const saveMutation = useMutation({
    mutationFn: ({ path, content }: { path: string; content: string }) => {
      if (!sandboxId || !accessToken) {
        throw new Error("Missing sandbox or access token");
      }
      return saveFile(accessToken, sandboxId, path, content);
    },
    onSuccess: (_result, variables) => {
      if (!sandboxId) return;
      queryClient.setQueryData(
        sandboxFileKeys.content(sandboxId, variables.path),
        { path: variables.path, content: variables.content },
      );
      void queryClient.invalidateQueries({
        queryKey: sandboxFileKeys.allFiles(sandboxId),
      });
    },
  });

  function updateDraft(content: string) {
    if (!filePath) return;
    setDraft({ path: filePath, content });
  }

  function save() {
    if (!sandboxId || !filePath || !accessToken || !hasUnsavedChanges) {
      return Promise.resolve();
    }
    return saveMutation.mutateAsync({ path: filePath, content });
  }

  return {
    content,
    error: fileQuery.error ?? saveMutation.error,
    hasUnsavedChanges,
    isLoading: fileQuery.isLoading,
    isSaving: saveMutation.isPending,
    save,
    updateDraft,
  };
}