import { request } from "../../shared/api/client";

export type Workspace = {
  id: string;
  name: string;
  slug: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
};

export type WorkspaceMember = {
  id: string;
  workspace_id: string;
  user_id: string;
  user: {
    id: string;
    email: string;
    name: string;
  };
  role: "OWNER" | "ADMIN" | "MEMBER";
  created_at: string;
  updated_at: string;
};

export type WorkspaceInvite = {
  id: string;
  workspace_id: string;
  email: string;
  role: string;
  status: string;
  invited_by: string;
  invited_by_user: {
    id: string;
    email: string;
    name: string;
  };
  token_hash: string;
  expires_at: string;
  created_at: string;
  updated_at: string;
};

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

export function listMembers(accessToken: string, workspaceId: string) {
  return authenticatedRequest<WorkspaceMember[]>(
    `/workspaces/${workspaceId}/members`,
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
