-- +goose Up
-- +goose StatementBegin

-- Add unique constraint to prevent duplicate pending invites for same email/workspace
DO $$
DECLARE
	pending_status_id BIGINT;
BEGIN
	SELECT id
	INTO pending_status_id
	FROM tbl_enum
	WHERE category = 'WORKSPACE_INVITE_STATUS'
	  AND code = 'PENDING';

	EXECUTE format(
		'CREATE UNIQUE INDEX idx_unique_pending_invite
		 ON tbl_workspace_invite (workspace_id, email)
		 WHERE status_id = %L',
		pending_status_id
	);
END $$;

-- Add index on token_hash for faster lookups
CREATE INDEX idx_tbl_workspace_invite_token 
ON tbl_workspace_invite (token_hash);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_unique_pending_invite;
DROP INDEX IF EXISTS idx_tbl_workspace_invite_token;

-- +goose StatementEnd
