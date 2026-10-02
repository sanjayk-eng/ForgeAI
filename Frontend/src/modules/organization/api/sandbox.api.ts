import { authenticatedRequest } from "./sandbox-client";

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
  organization_id: string;
  status: SandboxStatus;
  container_id: string;
  container_name: string;
  volume_name: string;
  last_error?: string;
  created_at: string;
  updated_at: string;
}

export function getSandboxByProject(accessToken: string, projectId: string, ensureMissing = false) {
	const query = ensureMissing ? "?ensure=true" : "";
  return authenticatedRequest<Sandbox>(
    accessToken,
    `/projects/${projectId}/sandbox${query}`,
  );
}

