import { request } from "../../../shared/api/client";
import type { CreateProjectInput, Project, ResolvedRepository } from "../features/projects/types/project.types";

export type {
  CreateProjectInput,
  Project,
  ProjectRepository,
  ProjectStatus,
  ProjectType,
  SyncStatus,
  ProjectRepositoryInput,
  ResolvedRepository,
} from "../features/projects/types/project.types";

function authenticatedRequest<T>(
  accessToken: string,
  path: string,
  options: RequestInit = {},
) {
  return request<T>(path, {
    ...options,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(options.headers ?? {}),
    },
  });
}

export function listProjects(accessToken: string, workspaceId: string) {
  return authenticatedRequest<Project[]>(
    accessToken,
    `/workspaces/${workspaceId}/projects`,
  );
}

export function createProject(
  accessToken: string,
  workspaceId: string,
  input: CreateProjectInput,
) {
  return authenticatedRequest<Project>(
    accessToken,
    `/workspaces/${workspaceId}/projects`,
    { method: "POST", body: JSON.stringify(input) },
  );
}

export function syncProject(accessToken: string, projectId: string) {
  return authenticatedRequest<Project>(
    accessToken,
    `/projects/${projectId}/repository/sync`,
    { method: "POST" },
  );
}

export function resolveRepository(accessToken: string, repositoryUrl: string) {
  return authenticatedRequest<ResolvedRepository>(accessToken, "/projects/repository/resolve", {
    method: "POST",
    body: JSON.stringify({ repository_url: repositoryUrl }),
  });
}