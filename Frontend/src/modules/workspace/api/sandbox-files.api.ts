import { authenticatedRequest } from "./sandbox-client";

export interface FileEntry {
  name: string;
  path: string;
  is_directory: boolean;
  size?: number;
  modified_at?: string;
}

export function listFiles(accessToken: string, sandboxId: string, path: string = "/workspace") {
  return authenticatedRequest<{ files: FileEntry[] }>(
    accessToken,
    `/sandboxes/${sandboxId}/files?path=${encodeURIComponent(path)}`,
  );
}

export function readFile(accessToken: string, sandboxId: string, path: string) {
  return authenticatedRequest<{ content: string; path: string }>(
    accessToken,
    `/sandboxes/${sandboxId}/files/${encodeURIComponent(path)}`,
  );
}

export function saveFile(accessToken: string, sandboxId: string, path: string, content: string) {
  return authenticatedRequest<{ path: string }>(
    accessToken,
    `/sandboxes/${sandboxId}/files/save`,
    { method: "POST", body: JSON.stringify({ path, content }) },
  );
}

export function createFileOrDirectory(accessToken: string, sandboxId: string, path: string, type: "file" | "directory") {
  return authenticatedRequest<{ path: string }>(
    accessToken,
    `/sandboxes/${sandboxId}/files/create`,
    { method: "POST", body: JSON.stringify({ path, type }) },
  );
}

export function deletePath(accessToken: string, sandboxId: string, path: string) {
  return authenticatedRequest<{ path: string }>(
    accessToken,
    `/sandboxes/${sandboxId}/files`,
    { method: "DELETE", body: JSON.stringify({ path }) },
  );
}

export function renamePath(accessToken: string, sandboxId: string, oldPath: string, newPath: string) {
  return authenticatedRequest<{ old_path: string; new_path: string }>(
    accessToken,
    `/sandboxes/${sandboxId}/files/rename`,
    { method: "PATCH", body: JSON.stringify({ old_path: oldPath, new_path: newPath }) },
  );
}

export interface GoDefinition {
  path: string;
  line: number;
  column: number;
}

export function getGoDefinition(accessToken: string, sandboxId: string, path: string, line: number, column: number) {
  const payload = { path, line, column };
  console.log("[ForgeAI] sending Go definition request", {
    sandboxId,
    endpoint: `/sandboxes/${sandboxId}/definition`,
    payload,
  });

  return authenticatedRequest<GoDefinition>(
    accessToken,
    `/sandboxes/${sandboxId}/definition`,
    { method: "POST", body: JSON.stringify(payload) },
  );
}