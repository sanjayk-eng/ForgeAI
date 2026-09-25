-- ForgeAI PostgreSQL schema
--
-- This file represents the final schema produced by the ordered files in
-- Backend/migrations. Runtime deployments should continue using Goose.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE tbl_enum (
    id BIGSERIAL PRIMARY KEY,
    category VARCHAR(100) NOT NULL,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tbl_enum_category_code UNIQUE (category, code)
);

INSERT INTO tbl_enum (category, code, name)
VALUES
    ('AUTH_PROVIDER', 'GOOGLE', 'Google'),
    ('AUTH_PROVIDER', 'GITHUB', 'GitHub'),
    ('WORKSPACE_ROLE', 'OWNER', 'Owner'),
    ('WORKSPACE_ROLE', 'ADMIN', 'Admin'),
    ('WORKSPACE_ROLE', 'MEMBER', 'Member'),
    ('WORKSPACE_INVITE_STATUS', 'PENDING', 'Pending'),
    ('WORKSPACE_INVITE_STATUS', 'ACCEPTED', 'Accepted'),
    ('WORKSPACE_INVITE_STATUS', 'REJECTED', 'Rejected'),
    ('WORKSPACE_INVITE_STATUS', 'EXPIRED', 'Expired'),
    ('WORKSPACE_INVITE_STATUS', 'REVOKED', 'Revoked'),
    ('PROJECT_TYPE', 'REPOSITORY', 'Repository'),
    ('PROJECT_TYPE', 'EMPTY', 'Empty'),
    ('PROJECT_STATUS', 'ACTIVE', 'Active'),
    ('PROJECT_STATUS', 'ARCHIVED', 'Archived'),
    ('PROJECT_REPOSITORY_SYNC_STATUS', 'PENDING', 'Pending'),
    ('PROJECT_REPOSITORY_SYNC_STATUS', 'SYNCING', 'Syncing'),
    ('PROJECT_REPOSITORY_SYNC_STATUS', 'SYNCED', 'Synced'),
    ('PROJECT_REPOSITORY_SYNC_STATUS', 'FAILED', 'Failed');

CREATE TABLE tbl_user (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    password_hash TEXT,
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tbl_oauth_account (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES tbl_user(id) ON DELETE CASCADE,
    provider_id BIGINT NOT NULL REFERENCES tbl_enum(id),
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_oauth_provider_user UNIQUE (provider_id, provider_user_id)
);

CREATE TABLE tbl_workspace (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(150) NOT NULL UNIQUE,
    owner_id UUID NOT NULL REFERENCES tbl_user(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tbl_workspace_member (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES tbl_workspace(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES tbl_user(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES tbl_enum(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_tbl_workspace_member_workspace_user UNIQUE (workspace_id, user_id)
);

CREATE TABLE tbl_workspace_invite (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES tbl_workspace(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role_id BIGINT NOT NULL REFERENCES tbl_enum(id),
    status_id BIGINT NOT NULL REFERENCES tbl_enum(id),
    invited_by UUID NOT NULL REFERENCES tbl_user(id),
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tbl_workspace_invite_workspace
    ON tbl_workspace_invite (workspace_id);
CREATE INDEX idx_tbl_workspace_invite_email
    ON tbl_workspace_invite (email);
CREATE INDEX idx_tbl_workspace_invite_status
    ON tbl_workspace_invite (status_id);
CREATE INDEX idx_tbl_workspace_invite_token
    ON tbl_workspace_invite (token_hash);

DO $$
DECLARE
    pending_status_id BIGINT;
BEGIN
    SELECT id INTO pending_status_id
    FROM tbl_enum
    WHERE category = 'WORKSPACE_INVITE_STATUS' AND code = 'PENDING';

    EXECUTE format(
        'CREATE UNIQUE INDEX idx_unique_pending_invite
         ON tbl_workspace_invite (workspace_id, email)
         WHERE status_id = %L',
        pending_status_id
    );
END $$;

CREATE TABLE tbl_email_verification (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES tbl_user(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tbl_project (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(150) NOT NULL,
    description TEXT,
    type_id BIGINT NOT NULL,
    status_id BIGINT NOT NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_tbl_project_workspace
        FOREIGN KEY (workspace_id) REFERENCES tbl_workspace(id) ON DELETE CASCADE,
    CONSTRAINT fk_tbl_project_type
        FOREIGN KEY (type_id) REFERENCES tbl_enum(id),
    CONSTRAINT fk_tbl_project_status
        FOREIGN KEY (status_id) REFERENCES tbl_enum(id),
    CONSTRAINT fk_tbl_project_created_by
        FOREIGN KEY (created_by) REFERENCES tbl_user(id),
    CONSTRAINT uq_tbl_project_workspace_slug UNIQUE (workspace_id, slug),
    CONSTRAINT ck_tbl_project_name_not_empty CHECK (btrim(name) <> ''),
    CONSTRAINT ck_tbl_project_slug_not_empty CHECK (btrim(slug) <> '')
);

CREATE TABLE tbl_project_repository (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    github_repository_id BIGINT NOT NULL,
    github_owner VARCHAR(255) NOT NULL,
    github_repository_name VARCHAR(255) NOT NULL,
    repository_url TEXT NOT NULL,
    default_branch VARCHAR(255) NOT NULL,
    sync_status_id BIGINT NOT NULL,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_tbl_project_repository_project
        FOREIGN KEY (project_id) REFERENCES tbl_project(id) ON DELETE CASCADE,
    CONSTRAINT fk_tbl_project_repository_workspace
        FOREIGN KEY (workspace_id) REFERENCES tbl_workspace(id) ON DELETE CASCADE,
    CONSTRAINT fk_tbl_project_repository_sync_status
        FOREIGN KEY (sync_status_id) REFERENCES tbl_enum(id),
    CONSTRAINT uq_tbl_project_repository_project UNIQUE (project_id),
    CONSTRAINT uq_tbl_project_repository_workspace_github_repository
        UNIQUE (workspace_id, github_repository_id),
    CONSTRAINT ck_tbl_project_repository_github_owner_not_empty
        CHECK (btrim(github_owner) <> ''),
    CONSTRAINT ck_tbl_project_repository_name_not_empty
        CHECK (btrim(github_repository_name) <> ''),
    CONSTRAINT ck_tbl_project_repository_url_not_empty
        CHECK (btrim(repository_url) <> ''),
    CONSTRAINT ck_tbl_project_repository_default_branch_not_empty
        CHECK (btrim(default_branch) <> '')
);
