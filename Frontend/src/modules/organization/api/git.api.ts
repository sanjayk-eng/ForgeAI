import { authenticatedRequest } from "./sandbox-client";

export interface GitStatusPayload {
  branch: string;
  is_dirty: boolean;
  modified: string[];
  staged: string[];
  untracked: string[];
  ahead?: number;
  behind?: number;
}

export interface GitDiffEntry {
  path: string;
  old_path?: string;
  content: string;
  status: string;
  original_content?: string;
}

interface GitStatusResponse {
  status: GitStatusPayload;
}

interface GitDiffResponse {
  files: GitDiffEntry[];
}

export function getGitStatus(accessToken: string, sandboxId: string) {
  return authenticatedRequest<GitStatusResponse>(accessToken, `/sandboxes/${sandboxId}/git/status`);
}

export function getGitDiff(accessToken: string, sandboxId: string, path?: string) {
  const query = path ? `?path=${encodeURIComponent(path)}` : "";
  return authenticatedRequest<GitDiffResponse>(
    accessToken,
    `/sandboxes/${sandboxId}/git/diff${query}`,
  );
}

export function commitGitChanges(accessToken: string, sandboxId: string, message: string, files: string[]) {
  return authenticatedRequest<{ message: string; hash: string }>(
    accessToken,
    `/sandboxes/${sandboxId}/git/commit`,
    { method: "POST", body: JSON.stringify({ message, files }) },
  );
}

export function revertGitChanges(accessToken: string, sandboxId: string, files: string[], all = false) {
  return authenticatedRequest<{ message: string; status: string }>(
    accessToken,
    `/sandboxes/${sandboxId}/git/revert`,
    { method: "POST", body: JSON.stringify({ files, all }) },
  );
}

export function pushGitChanges(accessToken: string, sandboxId: string, remote?: string, branch?: string) {
  return authenticatedRequest<{ message: string; status: string }>(
    accessToken,
    `/sandboxes/${sandboxId}/git/push`,
    { method: "POST", body: JSON.stringify({ remote, branch }) },
  );
}