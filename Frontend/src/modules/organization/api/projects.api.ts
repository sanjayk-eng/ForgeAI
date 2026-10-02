import { request } from "../../../shared/api/client";
import { pageQuery, type PageParams, type PagedResult } from "../../../shared/api/pagination";
import type {
  CreateProjectInput,
  Project,
  ProjectRepositoryInput,
  ResolvedRepository,
  UpdateProjectInput,
} from "../features/projects/types/project.types";

export type {
  CreateProjectInput,
  Project,
  ProjectRepository,
  ProjectStatus,
  ProjectType,
  ProjectRepositoryInput,
  ResolvedRepository,
  UpdateProjectInput,
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

export function listProjects(accessToken: string, organizationId: string, params?: PageParams) {
  return authenticatedRequest<PagedResult<Project>>(
    accessToken,
    `/organizations/${organizationId}/projects${pageQuery(params)}`,
  );
}

export function createProject(
  accessToken: string,
  organizationId: string,
  input: CreateProjectInput,
) {
  return authenticatedRequest<Project>(
    accessToken,
    `/organizations/${organizationId}/projects`,
    { method: "POST", body: JSON.stringify(input) },
  );
}

export function resolveRepository(accessToken: string, repositoryUrl: string) {
  return authenticatedRequest<ResolvedRepository>(accessToken, "/projects/repository/resolve", {
    method: "POST",
    body: JSON.stringify({ repository_url: repositoryUrl }),
  });
}

export function connectRepository(
  accessToken: string,
  projectId: string,
  repository: ProjectRepositoryInput,
) {
  return authenticatedRequest<ProjectRepositoryInput>(
    accessToken,
    `/projects/${projectId}/repository`,
    { method: "POST", body: JSON.stringify(repository) },
  );
}

export function updateProject(accessToken: string, projectId: string, input: UpdateProjectInput) {
  return authenticatedRequest<Project>(accessToken, `/projects/${projectId}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function updateRepositoryBranch(accessToken: string, projectId: string, defaultBranch: string) {
  return authenticatedRequest<ProjectRepositoryInput>(accessToken, `/projects/${projectId}/repository/branch`, {
    method: "PATCH",
    body: JSON.stringify({ default_branch: defaultBranch }),
  });
}

export function deleteProject(accessToken: string, projectId: string) {
  return authenticatedRequest<null>(accessToken, `/projects/${projectId}`, {
    method: "DELETE",
  });
}