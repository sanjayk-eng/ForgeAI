-- +goose Up
CREATE TABLE tbl_github_installation_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    state_hash CHAR(64) NOT NULL UNIQUE,
    organization_id UUID NOT NULL
        REFERENCES tbl_organization(id) ON DELETE CASCADE,
    user_id UUID NOT NULL
        REFERENCES tbl_user(id) ON DELETE CASCADE,
    integration_id UUID
        REFERENCES tbl_organization_integration(id) ON DELETE SET NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_github_installation_state_expiry
    ON tbl_github_installation_state(expires_at);

-- +goose Down
DROP TABLE IF EXISTS tbl_github_installation_state;