import { request } from "../../../shared/api/client";
import type { Workspace } from "../types/workspace.types";

type AuthenticatedOptions = RequestInit & { accessToken: string };

function authenticatedRequest<T>(
  path: string,
  options: AuthenticatedOptions,
): Promise<T> {
  const { accessToken, ...requestOptions } = options;
  return request<T>(path, {
    ...requestOptions,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(requestOptions.headers ?? {}),
    },
  });
}

export function listWorkspaces(accessToken: string) {
  return authenticatedRequest<Workspace[]>("/workspaces", { accessToken });
}

export function createWorkspace(accessToken: string, name: string) {
  return authenticatedRequest<Workspace>("/workspaces", {
    accessToken,
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

export function updateWorkspace(
  accessToken: string,
  workspaceId: string,
  name: string,
) {
  return authenticatedRequest<Workspace>(`/workspaces/${workspaceId}`, {
    accessToken,
    method: "PATCH",
    body: JSON.stringify({ name }),
  });
}
