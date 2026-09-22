-- +goose Up

-- +goose StatementBegin

INSERT INTO tbl_enum (
    category,
    code,
    name
)
VALUES
    ('WORKSPACE_INVITE_STATUS', 'PENDING', 'Pending'),
    ('WORKSPACE_INVITE_STATUS', 'ACCEPTED', 'Accepted'),
    ('WORKSPACE_INVITE_STATUS', 'REJECTED', 'Rejected'),
    ('WORKSPACE_INVITE_STATUS', 'EXPIRED', 'Expired'),
    ('WORKSPACE_INVITE_STATUS', 'REVOKED', 'Revoked');

CREATE TABLE tbl_workspace_invite (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    workspace_id UUID NOT NULL
        REFERENCES tbl_workspace(id)
        ON DELETE CASCADE,

    email VARCHAR(255) NOT NULL,

    role_id BIGINT NOT NULL
        REFERENCES tbl_enum(id),

    status_id BIGINT NOT NULL
        REFERENCES tbl_enum(id),

    invited_by UUID NOT NULL
        REFERENCES tbl_user(id),

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

-- +goose StatementEnd


-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS tbl_workspace_invite;

DELETE FROM tbl_enum
WHERE category = 'WORKSPACE_INVITE_STATUS'
  AND code IN (
      'PENDING',
      'ACCEPTED',
      'REJECTED',
      'EXPIRED',
      'REVOKED'
  );

-- +goose StatementEnd