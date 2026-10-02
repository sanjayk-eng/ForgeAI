// Core organization types
export type Organization = {
  id: string;
  name: string;
  slug: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
};

export type OrganizationRole = "OWNER" | "ADMIN" | "MEMBER";

export type OrganizationMember = {
  id: string;
  organization_id: string;
  user_id: string;
  user: {
    id: string;
    email: string;
    name: string;
  };
  role: OrganizationRole;
  created_at: string;
  updated_at: string;
};

export type InviteStatus = "PENDING" | "ACCEPTED" | "REJECTED" | "EXPIRED" | "REVOKED";

export type OrganizationInvite = {
  id: string;
  organization_id: string;
  organization_name?: string;
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
