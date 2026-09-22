-- +goose Up

-- +goose StatementBegin

INSERT INTO tbl_enum (
    category,
    code,
    name
)
VALUES
    ('WORKSPACE_ROLE', 'OWNER', 'Owner'),
    ('WORKSPACE_ROLE', 'ADMIN', 'Admin'),
    ('WORKSPACE_ROLE', 'MEMBER', 'Member');

CREATE TABLE tbl_workspace_member (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    workspace_id UUID NOT NULL
        REFERENCES tbl_workspace(id)
        ON DELETE CASCADE,

    user_id UUID NOT NULL
        REFERENCES tbl_user(id)
        ON DELETE CASCADE,

    role_id BIGINT NOT NULL
        REFERENCES tbl_enum(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT uq_tbl_workspace_member_workspace_user
        UNIQUE (workspace_id, user_id)
);

-- +goose StatementEnd


-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS tbl_workspace_member;

DELETE FROM tbl_enum
WHERE category = 'WORKSPACE_ROLE'
  AND code IN ('OWNER', 'ADMIN', 'MEMBER');

-- +goose StatementEnd