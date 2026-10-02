import { authenticatedRequest } from "./sandbox-client";

export interface AgentStatus {
  configured: boolean;
  model?: string;
}

export interface AgentTaskResult {
  message: string;
  changed_files: string[];
}

export function getAgentStatus(accessToken: string, sandboxId: string) {
  return authenticatedRequest<AgentStatus>(accessToken, `/sandboxes/${sandboxId}/agent/status`);
}

export function runAgentTask(accessToken: string, sandboxId: string, prompt: string) {
  return authenticatedRequest<AgentTaskResult>(
    accessToken,
    `/sandboxes/${sandboxId}/agent/tasks`,
    { method: "POST", body: JSON.stringify({ prompt }) },
  );
}