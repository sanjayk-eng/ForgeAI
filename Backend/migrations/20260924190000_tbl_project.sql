-- +goose Up

-- +goose StatementBegin

INSERT INTO tbl_enum (
    category,
    code,
    name
)
VALUES
    ('PROJECT_TYPE', 'REPOSITORY', 'Repository'),
    ('PROJECT_TYPE', 'EMPTY', 'Empty'),
    ('PROJECT_STATUS', 'ACTIVE', 'Active'),
    ('PROJECT_STATUS', 'ARCHIVED', 'Archived'),
    ('PROJECT_REPOSITORY_SYNC_STATUS', 'PENDING', 'Pending'),
    ('PROJECT_REPOSITORY_SYNC_STATUS', 'SYNCING', 'Syncing'),
    ('PROJECT_REPOSITORY_SYNC_STATUS', 'SYNCED', 'Synced'),
    ('PROJECT_REPOSITORY_SYNC_STATUS', 'FAILED', 'Failed');

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
        FOREIGN KEY (workspace_id)
        REFERENCES tbl_workspace(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_tbl_project_type
        FOREIGN KEY (type_id)
        REFERENCES tbl_enum(id),

    CONSTRAINT fk_tbl_project_status
        FOREIGN KEY (status_id)
        REFERENCES tbl_enum(id),

    CONSTRAINT fk_tbl_project_created_by
        FOREIGN KEY (created_by)
        REFERENCES tbl_user(id),

    CONSTRAINT uq_tbl_project_workspace_slug
        UNIQUE (workspace_id, slug),

    CONSTRAINT ck_tbl_project_name_not_empty
        CHECK (btrim(name) <> ''),

    CONSTRAINT ck_tbl_project_slug_not_empty
        CHECK (btrim(slug) <> '')
);

CREATE TABLE tbl_project_repository (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    project_id UUID NOT NULL,
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
        FOREIGN KEY (project_id)
        REFERENCES tbl_project(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_tbl_project_repository_sync_status
        FOREIGN KEY (sync_status_id)
        REFERENCES tbl_enum(id),

    CONSTRAINT uq_tbl_project_repository_project
        UNIQUE (project_id),

    CONSTRAINT uq_tbl_project_repository_github_repository
        UNIQUE (github_repository_id),

    CONSTRAINT ck_tbl_project_repository_github_owner_not_empty
        CHECK (btrim(github_owner) <> ''),

    CONSTRAINT ck_tbl_project_repository_name_not_empty
        CHECK (btrim(github_repository_name) <> ''),

    CONSTRAINT ck_tbl_project_repository_url_not_empty
        CHECK (btrim(repository_url) <> ''),

    CONSTRAINT ck_tbl_project_repository_default_branch_not_empty
        CHECK (btrim(default_branch) <> '')
);

-- +goose StatementEnd


-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS tbl_project_repository;
DROP TABLE IF EXISTS tbl_project;

DELETE FROM tbl_enum
WHERE category = 'PROJECT_TYPE'
  AND code IN ('REPOSITORY', 'EMPTY');

DELETE FROM tbl_enum
WHERE category = 'PROJECT_STATUS'
  AND code IN ('ACTIVE', 'ARCHIVED');

DELETE FROM tbl_enum
WHERE category = 'PROJECT_REPOSITORY_SYNC_STATUS'
  AND code IN ('PENDING', 'SYNCING', 'SYNCED', 'FAILED');

-- +goose StatementEnd