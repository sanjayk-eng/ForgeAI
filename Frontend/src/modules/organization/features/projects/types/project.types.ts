export type ProjectStatus = "ACTIVE" | "ARCHIVED";
export type SyncStatus = "PENDING" | "SYNCING" | "SYNCED" | "FAILED";

export type ProjectRepository = {
  id: string;
  project_id: string;
  github_repository_id: number;
  github_owner: string;
  github_repository_name: string;
  repository_url: string;
  private: boolean;
  default_branch: string;
  sync_status: SyncStatus;
  last_synced_at?: string;
  updated_at: string;
};

export type Project = {
  id: string;
  organization_id: string;
  name: string;
  slug: string;
  description?: string;
  status: ProjectStatus;
  created_at: string;
  updated_at: string;
  repository?: ProjectRepository;
};

export type CreateProjectInput = {
  name: string;
  description?: string;
};

export type ProjectRepositoryInput = {
  github_repository_id: number;
  github_owner: string;
  github_repository_name: string;
  repository_url: string;
  private: boolean;
  default_branch: string;
};

export type ResolvedRepository = {
  repository: ProjectRepositoryInput;
  branches: string[];
};

export type UpdateProjectInput = {
  name: string;
  description?: string;
  status?: ProjectStatus;
};