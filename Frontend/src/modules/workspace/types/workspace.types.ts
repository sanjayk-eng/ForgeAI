// Core workspace types
export type Workspace = {
  id: string;
  name: string;
  slug: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
};

export type WorkspaceRole = "OWNER" | "ADMIN" | "MEMBER";

export type WorkspaceMember = {
  id: string;
  workspace_id: string;
  user_id: string;
  user: {
    id: string;
    email: string;
    name: string;
  };
  role: WorkspaceRole;
  created_at: string;
  updated_at: string;
};

export type InviteStatus = "PENDING" | "ACCEPTED" | "REJECTED" | "EXPIRED" | "REVOKED";

export type WorkspaceInvite = {
  id: string;
  workspace_id: string;
  workspace_name?: string;
  email: string;
  role: string;
  status: InviteStatus;
  invited_by: string;
  invited_by_user?: {
    id: string;
    email: string;
    name: string;
  };
  token?: string;
  expires_at: string;
  created_at: string;
  updated_at: string;
};
