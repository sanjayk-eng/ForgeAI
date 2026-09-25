-- +goose Up

-- +goose StatementBegin

ALTER TABLE tbl_project_repository
    ADD COLUMN workspace_id UUID;

UPDATE tbl_project_repository pr
SET workspace_id = p.workspace_id
FROM tbl_project p
WHERE p.id = pr.project_id;

ALTER TABLE tbl_project_repository
    ALTER COLUMN workspace_id SET NOT NULL;

ALTER TABLE tbl_project_repository
    ADD CONSTRAINT fk_tbl_project_repository_workspace
        FOREIGN KEY (workspace_id)
        REFERENCES tbl_workspace(id)
        ON DELETE CASCADE;

ALTER TABLE tbl_project_repository
    DROP CONSTRAINT uq_tbl_project_repository_github_repository;

ALTER TABLE tbl_project_repository
    ADD CONSTRAINT uq_tbl_project_repository_workspace_github_repository
        UNIQUE (workspace_id, github_repository_id);

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin

ALTER TABLE tbl_project_repository
    DROP CONSTRAINT uq_tbl_project_repository_workspace_github_repository;

ALTER TABLE tbl_project_repository
    ADD CONSTRAINT uq_tbl_project_repository_github_repository
        UNIQUE (github_repository_id);

ALTER TABLE tbl_project_repository
    DROP CONSTRAINT fk_tbl_project_repository_workspace;

ALTER TABLE tbl_project_repository
    DROP COLUMN workspace_id;

-- +goose StatementEnd
