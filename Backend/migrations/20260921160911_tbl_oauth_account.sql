-- +goose Up

-- +goose StatementBegin

CREATE TABLE tbl_oauth_account (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES tbl_user(id)
        ON DELETE CASCADE,

    provider_id BIGINT NOT NULL
        REFERENCES tbl_enum(id),

    provider_user_id VARCHAR(255) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_oauth_provider_user
        UNIQUE (provider_id, provider_user_id)
);

-- +goose StatementEnd


-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS tbl_oauth_account;

-- +goose StatementEnd