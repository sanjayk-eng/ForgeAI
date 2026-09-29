import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { getGitDiff } from "../../../../api/git.api";
import { readFile, saveFile } from "../../../../api/sandbox-files.api";
import { sandboxFileKeys } from "./fileTree";

interface FileDraft {
  path: string;
  content: string;
}

export interface FileContentDiff {
  original: string;
  modified: string;
}

export function useFileEditor(
  sandboxId: string | null,
  filePath: string | null,
  accessToken: string | null,
) {
  const queryClient = useQueryClient();
  const [draft, setDraft] = useState<FileDraft | null>(null);
  const [diff, setDiff] = useState<FileContentDiff | null>(null);
  const [externalChange, setExternalChange] = useState(false);
  const initialGitDiffPath = useRef<string | null>(null);

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

  useEffect(() => {
    initialGitDiffPath.current = null;
    setDiff(null);
    setExternalChange(false);
  }, [filePath]);

  useEffect(() => {
    if (!filePath || !fileQuery.data || initialGitDiffPath.current === filePath) return;
    initialGitDiffPath.current = filePath;
    let active = true;
    const currentContent = fileQuery.data.content;
    void gitOriginalContent(filePath).then((original) => {
      if (active && original !== undefined && original !== currentContent) {
        setDiff({ original, modified: currentContent });
      }
    });
    return () => {
      active = false;
    };
  }, [accessToken, filePath, fileQuery.data?.content, sandboxId]);

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
    const original = fileQuery.data?.content ?? "";
    return saveMutation.mutateAsync({ path: filePath, content }).then(() => {
      if (original !== content) {
        void gitOriginalContent(filePath).then((gitOriginal) => {
          setDiff({ original: gitOriginal ?? original, modified: content });
        });
      }
      setExternalChange(false);
    });
  }

  async function refreshWithDiff() {
    if (!filePath) return;
    const original = fileQuery.data?.content ?? "";
    const hadUnsavedDraft = Boolean(draft?.path === filePath && draft.content !== original);
    const result = await fileQuery.refetch();
    const modified = result.data?.content;
    if (modified === undefined || modified === original) return;
    const gitOriginal = await gitOriginalContent(filePath);
    setDiff({ original: gitOriginal ?? original, modified });
    if (hadUnsavedDraft) {
      setExternalChange(true);
      return;
    }
    setDraft(null);
    setExternalChange(false);
  }

  async function gitOriginalContent(path: string) {
    if (!accessToken || !sandboxId) return undefined;
    const relativePath = path.replaceAll("\\", "/").replace(/^\/workspace\//, "");
    try {
      const result = await getGitDiff(accessToken, sandboxId, relativePath);
      return result.files.find((file) => file.path === relativePath)?.original_content;
    } catch {
      return undefined;
    }
  }

  function useExternalVersion() {
    setDraft(null);
    setExternalChange(false);
  }

  return {
    content,
    error: fileQuery.error ?? saveMutation.error,
    hasUnsavedChanges,
    isLoading: fileQuery.isLoading,
    isSaving: saveMutation.isPending,
    diff,
    externalChange,
    save,
    refreshWithDiff,
    useExternalVersion,
    updateDraft,
  };
}