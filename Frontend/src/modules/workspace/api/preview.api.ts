import { authenticatedRequest } from "./sandbox-client";

export type PreviewInfo = {
  status: "starting" | "running" | "stopped" | "sandbox_unavailable" | "application_unavailable";
  port: number | null;
  url: string | null;
};

export function getProjectPreview(accessToken: string, projectId: string) {
  return authenticatedRequest<PreviewInfo>(
    accessToken,
    `/projects/${encodeURIComponent(projectId)}/preview`,
  );
}