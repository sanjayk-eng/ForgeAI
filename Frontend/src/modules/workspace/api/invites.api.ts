import { request } from "../../../shared/api/client";
import type { WorkspaceInvite } from "../types/workspace.types";

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

export function createInvite(
  accessToken: string,
  workspaceId: string,
  email: string,
  role: string,
) {
  return authenticatedRequest<WorkspaceInvite>(
    `/workspaces/${workspaceId}/invites`,
    {
      accessToken,
      method: "POST",
      body: JSON.stringify({ email, role }),
    },
  );
}

export function listInvites(accessToken: string, workspaceId: string) {
  return authenticatedRequest<WorkspaceInvite[]>(
    `/workspaces/${workspaceId}/invites`,
    { accessToken },
  );
}

export function getInviteByToken(token: string) {
  return request<WorkspaceInvite>(`/public/invites/${token}`);
}

export function acceptInvite(accessToken: string, token: string, email: string) {
  return authenticatedRequest<WorkspaceInvite>(`/invites/${token}/accept`, {
    accessToken,
    method: "POST",
    body: JSON.stringify({ email }),
  });
}

export function rejectInvite(token: string, email: string) {
  return request<WorkspaceInvite>(`/public/invites/${token}/reject`, {
    method: "POST",
    body: JSON.stringify({ email }),
  });
}

export function revokeInvite(
  accessToken: string,
  workspaceID: string,
  inviteID: string,
) {
  return authenticatedRequest<null>(
    `/workspaces/${workspaceID}/invites/${inviteID}`,
    {
      accessToken,
      method: "DELETE",
    },
  );
}
