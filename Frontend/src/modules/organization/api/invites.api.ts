import { request } from "../../../shared/api/client";
import type { OrganizationInvite } from "../types/organization.types";

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
  organizationId: string,
  email: string,
  role: string,
) {
  return authenticatedRequest<OrganizationInvite>(
    `/organizations/${organizationId}/invites`,
    {
      accessToken,
      method: "POST",
      body: JSON.stringify({ email, role }),
    },
  );
}

export function listInvites(accessToken: string, organizationId: string) {
  return authenticatedRequest<OrganizationInvite[]>(
    `/organizations/${organizationId}/invites`,
    { accessToken },
  );
}

export function getInviteByToken(token: string) {
  return request<OrganizationInvite>(`/public/invites/${token}`);
}

export function acceptInvite(accessToken: string, token: string, email: string) {
  return authenticatedRequest<OrganizationInvite>(`/invites/${token}/accept`, {
    accessToken,
    method: "POST",
    body: JSON.stringify({ email }),
  });
}

export function rejectInvite(token: string, email: string) {
  return request<OrganizationInvite>(`/public/invites/${token}/reject`, {
    method: "POST",
    body: JSON.stringify({ email }),
  });
}

export function revokeInvite(
  accessToken: string,
  organizationId: string,
  inviteID: string,
) {
  return authenticatedRequest<null>(
    `/organizations/${organizationId}/invites/${inviteID}`,
    {
      accessToken,
      method: "DELETE",
    },
  );
}

export function getMyPendingInvites(accessToken: string) {
  return authenticatedRequest<OrganizationInvite[]>(`/invites/my-pending`, {
    accessToken,
  });
}
