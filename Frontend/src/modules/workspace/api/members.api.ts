import { request } from "../../../shared/api/client";
import { pageQuery, type PageParams, type PagedResult } from "../../../shared/api/pagination";
import type { WorkspaceMember } from "../types/workspace.types";

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

export function listMembers(accessToken: string, workspaceId: string, params?: PageParams) {
  return authenticatedRequest<PagedResult<WorkspaceMember>>(
    `/workspaces/${workspaceId}/members${pageQuery(params)}`,
    { accessToken },
  );
}

export function updateMemberRole(
  accessToken: string,
  workspaceId: string,
  userId: string,
  role: string,
) {
  return authenticatedRequest<null>(
    `/workspaces/${workspaceId}/members/${userId}/role`,
    {
      accessToken,
      method: "PATCH",
      body: JSON.stringify({ role }),
    },
  );
}

export function removeMember(
  accessToken: string,
  workspaceId: string,
  userId: string,
) {
  return authenticatedRequest<null>(
    `/workspaces/${workspaceId}/members/${userId}`,
    {
      accessToken,
      method: "DELETE",
    },
  );
}
