-- +goose Up
-- =========================================================
-- ForgeAI: Organization-Based Architecture
-- PostgreSQL Initial Schema
-- =========================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =========================================================
-- 1. USER MANAGEMENT
-- =========================================================

CREATE TABLE tbl_user (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    password_hash TEXT,
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_email ON tbl_user(email);

-- =========================================================
-- 2. EMAIL VERIFICATION
-- =========================================================

CREATE TABLE tbl_email_verification (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES tbl_user(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_verification_user ON tbl_email_verification(user_id);
CREATE INDEX idx_email_verification_token ON tbl_email_verification(token_hash) WHERE verified_at IS NULL;

-- =========================================================
-- 3. INTEGRATION PROVIDER
-- =========================================================

CREATE TABLE tbl_integration_provider (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    auth_type VARCHAR(30) NOT NULL
        CHECK (auth_type IN ('GITHUB_APP', 'OAUTH', 'API_KEY')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tbl_integration_provider (code, name, auth_type)
VALUES
    ('GITHUB', 'GitHub', 'GITHUB_APP'),
    ('GITLAB', 'GitLab', 'OAUTH'),
    ('BITBUCKET', 'Bitbucket', 'OAUTH');

-- =========================================================
-- 4. OAUTH ACCOUNT (User-level GitHub/Google auth)
-- =========================================================

CREATE TABLE tbl_oauth_account (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES tbl_user(id) ON DELETE CASCADE,
    provider_code VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_oauth_provider_user UNIQUE (provider_code, provider_user_id)
);

CREATE INDEX idx_oauth_account_user ON tbl_oauth_account(user_id);

-- =========================================================
-- 5. ORGANIZATION
-- =========================================================

CREATE TABLE tbl_organization (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(180) NOT NULL UNIQUE,
    description TEXT,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    created_by UUID NOT NULL
        REFERENCES tbl_user(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_organization_created_by ON tbl_organization(created_by);
CREATE INDEX idx_organization_slug ON tbl_organization(slug) WHERE deleted_at IS NULL;

-- =========================================================
-- 6. ORGANIZATION MEMBER
-- =========================================================

CREATE TABLE tbl_organization_member (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL
        REFERENCES tbl_organization(id) ON DELETE CASCADE,
    user_id UUID NOT NULL
        REFERENCES tbl_user(id) ON DELETE RESTRICT,
    role VARCHAR(20) NOT NULL
        CHECK (role IN ('OWNER', 'ADMIN', 'MEMBER')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_organization_member_active
    ON tbl_organization_member(organization_id, user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_organization_member_user
    ON tbl_organization_member(user_id)
    WHERE deleted_at IS NULL;

-- =========================================================
-- 7. ORGANIZATION INVITE
-- =========================================================

CREATE TABLE tbl_organization_invite (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL
        REFERENCES tbl_organization(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'MEMBER'
        CHECK (role IN ('ADMIN', 'MEMBER')),
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING','ACCEPTED','REJECTED','EXPIRED','REVOKED')),
    invited_by UUID NOT NULL
        REFERENCES tbl_user(id) ON DELETE RESTRICT,
    accepted_by UUID
        REFERENCES tbl_user(id) ON DELETE SET NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_organization_invite_pending
    ON tbl_organization_invite(organization_id, LOWER(email))
    WHERE status = 'PENDING';

CREATE INDEX idx_organization_invite_status
    ON tbl_organization_invite(organization_id, status);

-- =========================================================
-- 8. ORGANIZATION INTEGRATION
-- =========================================================

CREATE TABLE tbl_organization_integration (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL
        REFERENCES tbl_organization(id) ON DELETE CASCADE,
    provider_id BIGINT NOT NULL
        REFERENCES tbl_integration_provider(id) ON DELETE RESTRICT,
    external_account_id VARCHAR(255) NOT NULL,
    external_account_name VARCHAR(255) NOT NULL,
    external_account_type VARCHAR(30),
    installation_id VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE','INACTIVE','EXPIRED','REVOKED')),
    connected_by UUID NOT NULL
        REFERENCES tbl_user(id) ON DELETE RESTRICT,
    connected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    disconnected_at TIMESTAMPTZ,
    provider_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_organization_integration_account
        UNIQUE (organization_id, provider_id, external_account_id),
    CONSTRAINT uq_organization_integration_id_org
        UNIQUE (id, organization_id)
);

CREATE UNIQUE INDEX uq_github_installation
    ON tbl_organization_integration(installation_id)
    WHERE installation_id IS NOT NULL;

CREATE INDEX idx_organization_integration_org
    ON tbl_organization_integration(organization_id);

CREATE INDEX idx_organization_integration_provider
    ON tbl_organization_integration(provider_id, status);

-- =========================================================
-- 9. PROJECT (Manual creation only, no import)
-- =========================================================

CREATE TABLE tbl_project (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL
        REFERENCES tbl_organization(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(180) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    created_by UUID NOT NULL
        REFERENCES tbl_user(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_project_organization_slug
        UNIQUE (organization_id, slug),
    CONSTRAINT uq_project_id_organization
        UNIQUE (id, organization_id)
);

CREATE INDEX idx_project_organization
    ON tbl_project(organization_id)
    WHERE deleted_at IS NULL;

-- =========================================================
-- 10. PROJECT REPOSITORY (Optional linking)
-- =========================================================

CREATE TABLE tbl_project_repository (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    project_id UUID NOT NULL,
    integration_id UUID NOT NULL,
    external_repository_id VARCHAR(255) NOT NULL,
    repository_name VARCHAR(255) NOT NULL,
    full_name VARCHAR(500) NOT NULL,
    default_branch VARCHAR(255),
    repository_url TEXT,
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    linked_by UUID NOT NULL
        REFERENCES tbl_user(id) ON DELETE RESTRICT,
    linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_project_repository_project
        UNIQUE (project_id),
    CONSTRAINT uq_project_repository_external
        UNIQUE (integration_id, external_repository_id),
    CONSTRAINT fk_project_repository_project
        FOREIGN KEY (project_id, organization_id)
        REFERENCES tbl_project(id, organization_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_project_repository_integration
        FOREIGN KEY (integration_id, organization_id)
        REFERENCES tbl_organization_integration(id, organization_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_project_repository_integration
    ON tbl_project_repository(integration_id);

-- =========================================================
-- 11. SANDBOX
-- =========================================================

CREATE TABLE tbl_sandbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL
        REFERENCES tbl_project(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'CREATING'
        CHECK (status IN ('CREATING','CREATED','STARTING','RUNNING',
                         'STOPPING','STOPPED','RESTARTING','FAILED',
                         'DESTROYING','DESTROYED')),
    container_id VARCHAR(255),
    container_name VARCHAR(255),
    volume_name VARCHAR(255),
    image VARCHAR(255),
    workspace_path VARCHAR(500) DEFAULT '/workspace',
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    destroyed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_sandbox_active_project
    ON tbl_sandbox(project_id)
    WHERE status NOT IN ('FAILED', 'DESTROYED');

CREATE INDEX idx_sandbox_status ON tbl_sandbox(status);

CREATE UNIQUE INDEX uq_sandbox_container_name
    ON tbl_sandbox(container_name)
    WHERE container_name IS NOT NULL AND status NOT IN ('DESTROYED');

CREATE UNIQUE INDEX uq_sandbox_volume_name
    ON tbl_sandbox(volume_name)
    WHERE volume_name IS NOT NULL AND status NOT IN ('DESTROYED');

-- =========================================================
-- 12. SANDBOX RESOURCE LIMIT
-- =========================================================

CREATE TABLE tbl_sandbox_resource_limit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sandbox_id UUID NOT NULL UNIQUE
        REFERENCES tbl_sandbox(id) ON DELETE CASCADE,
    cpu_quota INTEGER,
    memory_limit_mb INTEGER,
    pids_limit INTEGER,
    storage_limit_mb INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_sandbox_cpu CHECK (cpu_quota IS NULL OR cpu_quota > 0),
    CONSTRAINT chk_sandbox_memory CHECK (memory_limit_mb IS NULL OR memory_limit_mb > 0),
    CONSTRAINT chk_sandbox_storage CHECK (storage_limit_mb IS NULL OR storage_limit_mb > 0)
);

-- +goose Down
DROP TABLE IF EXISTS tbl_sandbox_resource_limit CASCADE;
DROP TABLE IF EXISTS tbl_sandbox CASCADE;
DROP TABLE IF EXISTS tbl_project_repository CASCADE;
DROP TABLE IF EXISTS tbl_project CASCADE;
DROP TABLE IF EXISTS tbl_organization_integration CASCADE;
DROP TABLE IF EXISTS tbl_organization_invite CASCADE;
DROP TABLE IF EXISTS tbl_organization_member CASCADE;
DROP TABLE IF EXISTS tbl_organization CASCADE;
DROP TABLE IF EXISTS tbl_oauth_account CASCADE;
DROP TABLE IF EXISTS tbl_integration_provider CASCADE;
DROP TABLE IF EXISTS tbl_email_verification CASCADE;
DROP TABLE IF EXISTS tbl_user CASCADE;
DROP EXTENSION IF EXISTS pgcrypto;
