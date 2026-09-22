-- +goose Up

-- +goose StatementBegin

CREATE TABLE tbl_workspace (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(150) NOT NULL,
    slug VARCHAR(150) NOT NULL UNIQUE,

    owner_id UUID NOT NULL
        REFERENCES tbl_user(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd


-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS tbl_workspace;

-- +goose StatementEnd