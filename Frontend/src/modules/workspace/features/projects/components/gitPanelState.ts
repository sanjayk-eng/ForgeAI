import type { GitDiffEntry, GitStatusPayload } from "../../../api/git.api";

export interface GitSelectionState {
  pathsKey: string;
  excluded: Set<string>;
}

export interface ChangedFile {
  path: string;
  status: string;
}

export function getGitSelectionSummary(changedPaths: string[], selection: GitSelectionState) {
  const excluded = selection.pathsKey === changedPathsKey(changedPaths) ? selection.excluded : new Set<string>();
  const selectedPaths = changedPaths.filter((path) => !excluded.has(path));

  return {
    excluded,
    selectedPaths,
    selectedCount: selectedPaths.length,
    totalCount: changedPaths.length,
  };
}

export function nextGitSelection(
  selection: GitSelectionState,
  changedPaths: string[],
  path: string,
  selected: boolean,
): GitSelectionState {
  const nextExcluded = new Set(selection.pathsKey === changedPathsKey(changedPaths) ? selection.excluded : []);
  if (selected) nextExcluded.delete(path);
  else nextExcluded.add(path);
  return { pathsKey: changedPathsKey(changedPaths), excluded: nextExcluded };
}

export function resetGitSelection(changedPaths: string[], mode: "all" | "none"): GitSelectionState {
  const nextExcluded = mode === "all" ? new Set<string>() : new Set(changedPaths);
  return { pathsKey: changedPathsKey(changedPaths), excluded: nextExcluded };
}

export function collectChangedFiles(status: GitStatusPayload | undefined, diffs: GitDiffEntry[]): ChangedFile[] {
  const files = new Map<string, string>();

  for (const file of diffs) {
    files.set(file.path, file.status);
  }

  for (const path of status?.staged ?? []) {
    if (!files.has(path)) files.set(path, "staged");
  }

  for (const path of status?.modified ?? []) {
    if (!files.has(path)) files.set(path, "modified");
  }

  for (const path of status?.untracked ?? []) {
    files.set(path, "untracked");
  }

  return [...files]
    .map(([path, fileStatus]) => ({ path, status: fileStatus }))
    .sort((left, right) => left.path.localeCompare(right.path));
}

export function buildCommitFileList(paths: string[], diffFiles: GitDiffEntry[]) {
  return [...new Set(
    paths.flatMap((filePath) => {
      const oldPath = diffFiles.find((file) => file.path === filePath)?.old_path;
      return oldPath ? [oldPath, filePath] : [filePath];
    }),
  )];
}

export function changedPathsKey(changedPaths: string[]) {
  return `${changedPaths.join("\0")}`;
}
