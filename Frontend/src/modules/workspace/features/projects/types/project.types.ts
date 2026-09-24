export type ProjectType = "REPOSITORY" | "EMPTY";
export type ProjectStatus = "ACTIVE" | "ARCHIVED";
export type SyncStatus = "PENDING" | "SYNCING" | "SYNCED" | "FAILED";

export type ProjectRepository = {
  id: string;
  project_id: string;
  github_repository_id: number;
  github_owner: string;
  github_repository_name: string;
  repository_url: string;
  default_branch: string;
  sync_status: SyncStatus;
  last_synced_at?: string;
};

export type Project = {
  id: string;
  workspace_id: string;
  name: string;
  slug: string;
  description?: string;
  type: ProjectType;
  status: ProjectStatus;
  created_at: string;
  updated_at: string;
  repository?: ProjectRepository;
};

export type CreateProjectInput = {
  name: string;
  description?: string;
  type: ProjectType;
  repository?: ProjectRepositoryInput;
};

export type ProjectRepositoryInput = {
  github_repository_id: number;
  github_owner: string;
  github_repository_name: string;
  repository_url: string;
  default_branch: string;
};

export type ResolvedRepository = {
  repository: ProjectRepositoryInput;
  branches: string[];
};