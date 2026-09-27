import { request } from "../../../shared/api/client";

export type SandboxStatus =
  | "CREATING"
  | "CREATED"
  | "STARTING"
  | "RUNNING"
  | "STOPPING"
  | "STOPPED"
  | "RESTARTING"
  | "FAILED"
  | "DESTROYING"
  | "DESTROYED";

export interface Sandbox {
  id: string;
  project_id: string;
  workspace_id: string;
  status: SandboxStatus;
  container_id: string;
  container_name: string;
  volume_name: string;
  last_error?: string;
  created_at: string;
  updated_at: string;
}

function authenticatedRequest<T>(
  accessToken: string,
  path: string,
  options: RequestInit = {},
) {
  return request<T>(path, {
    ...options,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(options.headers ?? {}),
    },
  });
}

export function getSandboxByProject(accessToken: string, projectId: string, ensureMissing = false) {
	const query = ensureMissing ? "?ensure=true" : "";
  return authenticatedRequest<Sandbox>(
    accessToken,
    `/projects/${projectId}/sandbox${query}`,
  );
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

export interface AgentStatus {
  configured: boolean;
  model?: string;
}

export interface AgentTaskResult {
  message: string;
  changed_files: string[];
}

export function getAgentStatus(accessToken: string, sandboxId: string) {
  return authenticatedRequest<AgentStatus>(
    accessToken,
    `/sandboxes/${sandboxId}/agent/status`,
  );
}

export function runAgentTask(accessToken: string, sandboxId: string, prompt: string) {
  return authenticatedRequest<AgentTaskResult>(
    accessToken,
    `/sandboxes/${sandboxId}/agent/tasks`,
    { method: "POST", body: JSON.stringify({ prompt }) },
  );
}

export interface FileEntry {
  name: string;
  path: string;
  is_directory: boolean;
  size?: number;
  modified_at?: string;
}
