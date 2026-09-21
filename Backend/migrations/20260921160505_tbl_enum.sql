-- +goose Up

-- +goose StatementBegin

CREATE TABLE tbl_enum (
    id BIGSERIAL PRIMARY KEY,

    category VARCHAR(100) NOT NULL,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(150) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_tbl_enum_category_code
        UNIQUE (category, code)
);

INSERT INTO tbl_enum (
    category,
    code,
    name
)
VALUES
    ('AUTH_PROVIDER', 'GOOGLE', 'Google'),
    ('AUTH_PROVIDER', 'GITHUB', 'GitHub');

-- +goose StatementEnd


-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS tbl_enum;

-- +goose StatementEnd