import { request } from "../../../shared/api/client";
import { pageQuery, type PageParams, type PagedResult } from "../../../shared/api/pagination";
import type {
  CreateProjectInput,
  GitHubRepositoryCatalog,
  GitHubRepositoryOption,
  Project,
  ProjectRepositoryInput,
  ResolvedRepository,
  UpdateProjectInput,
} from "../features/projects/types/project.types";

export type {
  CreateProjectInput,
  GitHubRepositoryCatalog,
  GitHubRepositoryOption,
  Project,
  ProjectRepository,
  ProjectStatus,
  ProjectType,
  SyncStatus,
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

export function listProjects(accessToken: string, workspaceId: string, params?: PageParams) {
  return authenticatedRequest<PagedResult<Project>>(
    accessToken,
    `/workspaces/${workspaceId}/projects${pageQuery(params)}`,
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

export function listGitHubRepositories(accessToken: string, workspaceId: string) {
  return authenticatedRequest<GitHubRepositoryCatalog>(
    accessToken,
    `/workspaces/${workspaceId}/github/repositories`,
  );
}

export function importGitHubRepositories(
  accessToken: string,
  workspaceId: string,
  repositories: GitHubRepositoryOption[],
) {
  return authenticatedRequest<{ projects: Project[]; skipped: number }>(
    accessToken,
    `/workspaces/${workspaceId}/github/repositories/import`,
    {
      method: "POST",
      body: JSON.stringify({ repositories }),
    },
  );
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

export function syncAllProjects(accessToken: string, workspaceId: string) {
  return authenticatedRequest<{ triggered: number; failed: number }>(
    accessToken,
    `/workspaces/${workspaceId}/projects/sync`,
    { method: "POST" },
  );
}

export function deleteProject(accessToken: string, projectId: string) {
  return authenticatedRequest<null>(accessToken, `/projects/${projectId}`, {
    method: "DELETE",
  });
}