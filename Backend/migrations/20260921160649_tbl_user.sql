-- +goose Up

-- +goose StatementBegin

CREATE TABLE tbl_user (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,

    password_hash TEXT,

    email_verified_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd


-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS tbl_user;

-- +goose StatementEnd