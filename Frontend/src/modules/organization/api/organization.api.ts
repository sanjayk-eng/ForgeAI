import { request } from "../../../shared/api/client";
import type { Organization } from "../types/organization.types";

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

export function listOrganizations(accessToken: string) {
  return authenticatedRequest<Organization[]>("/organizations", { accessToken });
}

export function createOrganization(accessToken: string, name: string) {
  return authenticatedRequest<Organization>("/organizations", {
    accessToken,
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

export function updateOrganization(
  accessToken: string,
  organizationId: string,
  name: string,
) {
  return authenticatedRequest<Organization>(`/organizations/${organizationId}`, {
    accessToken,
    method: "PATCH",
    body: JSON.stringify({ name }),
  });
}
