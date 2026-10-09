import axios from "axios";
import { apiUrl } from "../../../shared/api/client";

export type OrganizationIntegration = {
  id: string;
  provider_code: string;
  external_account_id: string;
  external_account_name: string;
  external_account_type: string;
  installation_id: string;
  status: "ACTIVE" | "INACTIVE" | "EXPIRED" | "REVOKED";
  connected_at: string;
};

export type ProjectGitHubExport = {
  project_id: string;
  integration_id: string;
  repository_id: string;
  repository_name: string;
  full_name: string;
  default_branch: string;
  repository_url: string;
  private: boolean;
};

type ApiEnvelope<T> = { data: T };

export async function listOrganizationIntegrations(
  accessToken: string,
  organizationId: string,
) {
  const response = await axios.get<ApiEnvelope<OrganizationIntegration[]>>(
    apiUrl(`/organizations/${encodeURIComponent(organizationId)}/integrations/github`),
    { headers: { Authorization: `Bearer ${accessToken}` } },
  );
  return response.data.data;
}

export async function startGitHubInstallation(
  accessToken: string,
  organizationId: string,
) {
  const response = await axios.get<ApiEnvelope<{ url: string }>>(
    apiUrl(`/integrations/github/install?organization_id=${encodeURIComponent(organizationId)}`),
    { headers: { Authorization: `Bearer ${accessToken}` } },
  );
  return response.data.data.url;
}

export async function disconnectGitHubIntegration(
  accessToken: string,
  organizationId: string,
  integrationId: string,
) {
  await axios.delete(
    apiUrl(`/organizations/${encodeURIComponent(organizationId)}/integrations/github/${encodeURIComponent(integrationId)}`),
    { headers: { Authorization: `Bearer ${accessToken}` } },
  );
}

export async function getProjectGitHubExport(
  accessToken: string,
  organizationId: string,
  projectId: string,
) {
  try {
    const response = await axios.get<ApiEnvelope<ProjectGitHubExport>>(
      apiUrl(`/organizations/${encodeURIComponent(organizationId)}/projects/${encodeURIComponent(projectId)}/github-export`),
      { headers: { Authorization: `Bearer ${accessToken}` } },
    );
    return response.data.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 404) return null;
    throw error;
  }
}

export async function createProjectGitHubExport(
  accessToken: string,
  organizationId: string,
  projectId: string,
  isPrivate: boolean,
) {
  try {
    const response = await axios.post<ApiEnvelope<ProjectGitHubExport>>(
      apiUrl(`/organizations/${encodeURIComponent(organizationId)}/projects/${encodeURIComponent(projectId)}/github-export`),
      { private: isPrivate },
      { headers: { Authorization: `Bearer ${accessToken}` } },
    );
    return response.data.data;
  } catch (error) {
    if (axios.isAxiosError<{ error?: { message?: string } }>(error)) {
      const message = error.response?.data?.error?.message;
      if (message) throw new Error(message);
    }
    throw error;
  }
}