-- +goose Up

-- +goose StatementBegin

INSERT INTO tbl_enum (category, code, name)
VALUES
    ('SANDBOX_STATUS', 'CREATING', 'Creating'),
    ('SANDBOX_STATUS', 'CREATED', 'Created'),
    ('SANDBOX_STATUS', 'STARTING', 'Starting'),
    ('SANDBOX_STATUS', 'RUNNING', 'Running'),
    ('SANDBOX_STATUS', 'STOPPING', 'Stopping'),
    ('SANDBOX_STATUS', 'STOPPED', 'Stopped'),
    ('SANDBOX_STATUS', 'RESTARTING', 'Restarting'),
    ('SANDBOX_STATUS', 'FAILED', 'Failed'),
    ('SANDBOX_STATUS', 'DESTROYING', 'Destroying'),
    ('SANDBOX_STATUS', 'DESTROYED', 'Destroyed');

CREATE TABLE tbl_sandbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL
        REFERENCES tbl_workspace(id) ON DELETE CASCADE,
    project_id UUID NOT NULL
        REFERENCES tbl_project(id) ON DELETE CASCADE,
    status_id BIGINT NOT NULL REFERENCES tbl_enum(id),
    volume_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    resource_limit JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_error TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_tbl_sandbox_project_not_empty CHECK (project_id IS NOT NULL)
);

CREATE TABLE tbl_sandbox_container (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sandbox_id UUID NOT NULL UNIQUE
        REFERENCES tbl_sandbox(id) ON DELETE CASCADE,
    source_id VARCHAR(255) NOT NULL DEFAULT '',
    source_type VARCHAR(50) NOT NULL DEFAULT 'docker',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_tbl_sandbox_container_source_type_not_empty CHECK (btrim(source_type) <> '')
);

CREATE INDEX idx_tbl_sandbox_project ON tbl_sandbox (project_id);
CREATE INDEX idx_tbl_sandbox_workspace ON tbl_sandbox (workspace_id);
CREATE INDEX idx_tbl_sandbox_status ON tbl_sandbox (status_id);
CREATE INDEX idx_tbl_sandbox_container_sandbox ON tbl_sandbox_container (sandbox_id);

DO $$
DECLARE
    destroyed_status_id BIGINT;
BEGIN
    SELECT id INTO destroyed_status_id
    FROM tbl_enum
    WHERE category = 'SANDBOX_STATUS' AND code = 'DESTROYED';

    EXECUTE format(
        'CREATE UNIQUE INDEX uq_tbl_active_sandbox_project
         ON tbl_sandbox (project_id)
         WHERE status_id <> %L',
        destroyed_status_id
    );
END $$;

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS tbl_sandbox_container;
DROP TABLE IF EXISTS tbl_sandbox;

DELETE FROM tbl_enum
WHERE category = 'SANDBOX_STATUS'
  AND code IN (
      'CREATING', 'CREATED', 'STARTING', 'RUNNING', 'STOPPING',
      'STOPPED', 'RESTARTING', 'FAILED', 'DESTROYING', 'DESTROYED'
  );

-- +goose StatementEnd
