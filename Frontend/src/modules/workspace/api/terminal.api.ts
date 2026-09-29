import { authenticatedRequest } from "./sandbox-client";

export type TerminalShell = {
  id: string;
  name: string;
  available: boolean;
  default: boolean;
};

export type TerminalSession = {
  session_id: string;
  project_id: string;
  sandbox_id: string;
  shell: string;
  cwd: string;
  status: string;
};

export async function listTerminalShells(accessToken: string, projectId: string) {
  const result = await authenticatedRequest<{ shells: TerminalShell[] }>(
    accessToken,
    `/projects/${encodeURIComponent(projectId)}/terminals/shells`,
  );
  return result.shells;
}

export function createTerminalSession(
  accessToken: string,
  projectId: string,
  shell: string,
  cols = 120,
  rows = 30,
) {
  return authenticatedRequest<TerminalSession>(
    accessToken,
    `/projects/${encodeURIComponent(projectId)}/terminals`,
    {
      method: "POST",
      body: JSON.stringify({ shell, cols, rows }),
    },
  );
}

export function closeTerminalSession(accessToken: string, projectId: string, sessionId: string) {
  return authenticatedRequest<{ session_id: string }>(
    accessToken,
    `/projects/${encodeURIComponent(projectId)}/terminals/${encodeURIComponent(sessionId)}`,
    { method: "DELETE" },
  );
}
