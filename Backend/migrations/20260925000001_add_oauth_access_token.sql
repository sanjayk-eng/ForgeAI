-- +goose Up

-- +goose StatementBegin

ALTER TABLE tbl_oauth_account
    ADD COLUMN access_token TEXT NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin

ALTER TABLE tbl_oauth_account
    DROP COLUMN access_token;

-- +goose StatementEnd