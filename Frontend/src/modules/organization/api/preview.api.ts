import { authenticatedRequest } from "./sandbox-client";

export type PreviewInfo = {
  status: "starting" | "running" | "stopped" | "sandbox_unavailable" | "application_unavailable";
  sandbox_id?: string;
  container_port: number | null;
  host_port?: number | null;
  protocol: "http" | "https";
  url: string | null;
};

export function getProjectPreview(accessToken: string, projectId: string) {
  return authenticatedRequest<PreviewInfo>(
    accessToken,
    `/projects/${encodeURIComponent(projectId)}/preview`,
  );
}