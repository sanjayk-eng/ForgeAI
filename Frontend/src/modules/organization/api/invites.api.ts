import { request } from "../../../shared/api/client";
import type { InviteStatus, OrganizationInvite } from "../types/organization.types";

export type InviteStatusUpdate = {
  organization_id?: string;
  status?: InviteStatus;
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

export function listInvites(accessToken: string, organizationId: string, status?: InviteStatus) {
  const query = status ? `?status=${encodeURIComponent(status)}` : "";
  return authenticatedRequest<OrganizationInvite[]>(
    `/organizations/${organizationId}/invites${query}`,
    { accessToken },
  );
}

export function getInviteByToken(token: string) {
  return request<OrganizationInvite>(`/public/invites/${token}`);
}

export function acceptInvite(accessToken: string, token: string, email: string) {
  return authenticatedRequest<InviteStatusUpdate>(`/invites/${token}/accept`, {
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

export function updateInviteStatus(
  accessToken: string,
  inviteID: string,
  status: "ACCEPTED" | "REJECTED" | "REVOKED",
) {
  return authenticatedRequest<InviteStatusUpdate>(`/invites/${encodeURIComponent(inviteID)}`, {
    accessToken,
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}

export async function getMyPendingInvites(accessToken: string) {
  const invites = await authenticatedRequest<OrganizationInvite[] | null>(`/invites/my-pending`, {
    accessToken,
  });
  return Array.isArray(invites) ? invites : [];
}
