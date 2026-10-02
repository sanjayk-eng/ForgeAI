import { request } from "../../../shared/api/client";
import { pageQuery, type PageParams, type PagedResult } from "../../../shared/api/pagination";
import type { OrganizationMember } from "../types/organization.types";

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

export function listMembers(accessToken: string, organizationId: string, params?: PageParams) {
  return authenticatedRequest<PagedResult<OrganizationMember>>(
    `/organizations/${organizationId}/members${pageQuery(params)}`,
    { accessToken },
  );
}

export function updateMemberRole(
  accessToken: string,
  organizationId: string,
  userId: string,
  role: string,
) {
  return authenticatedRequest<null>(
    `/organizations/${organizationId}/members/${userId}/role`,
    {
      accessToken,
      method: "PATCH",
      body: JSON.stringify({ role }),
    },
  );
}

export function removeMember(
  accessToken: string,
  organizationId: string,
  userId: string,
) {
  return authenticatedRequest<null>(
    `/organizations/${organizationId}/members/${userId}`,
    {
      accessToken,
      method: "DELETE",
    },
  );
}
