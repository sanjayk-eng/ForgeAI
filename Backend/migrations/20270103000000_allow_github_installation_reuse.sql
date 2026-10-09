-- +goose Up
DROP INDEX IF EXISTS uq_github_installation;

CREATE UNIQUE INDEX uq_github_installation_per_organization
    ON tbl_organization_integration(organization_id, installation_id)
    WHERE installation_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS uq_github_installation_per_organization;

CREATE UNIQUE INDEX uq_github_installation
    ON tbl_organization_integration(installation_id)
    WHERE installation_id IS NOT NULL;
