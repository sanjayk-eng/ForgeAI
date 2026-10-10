package github

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	appdatabase "ai-agent/pkg/database"

	"github.com/jmoiron/sqlx"
)

var (
	ErrStateInvalid             = errors.New("installation state is invalid or expired")
	ErrIntegrationNotFound      = errors.New("GitHub integration not found")
	ErrProjectNotFound          = errors.New("project not found")
	ErrExportRepositoryNotFound = errors.New("GitHub export repository not found")
)

type Repository interface {
	CreateState(ctx context.Context, stateHash, organizationID, userID string, expiresAt time.Time) error
	FindState(ctx context.Context, stateHash string) (State, error)
	CompleteInstallation(ctx context.Context, stateHash string, installation Installation) (Integration, error)
	ListGitHub(ctx context.Context, organizationID string) ([]Integration, error)
	FindExportContext(ctx context.Context, organizationID, projectID string) (ExportContext, error)
	SaveExportRepository(ctx context.Context, projectID, integrationID string, repository ExportRepository, userID string) (ExportRepository, error)
	FindExportRepository(ctx context.Context, organizationID, projectID string) (ExportRepository, error)
	FindExportCredentials(ctx context.Context, organizationID, projectID string) (string, string, error)
	DisconnectGitHub(ctx context.Context, organizationID, integrationID string) error
	SetInstallationStatus(ctx context.Context, installationID, status string) error
}

type repository struct {
	db *sqlx.DB
}

type installationState struct {
	organizationID        string
	userID                string
	consumedAt            sql.NullTime
	previousIntegrationID sql.NullString
}

type exportContextRow struct {
	ProjectID              string         `db:"project_id"`
	OrganizationID         string         `db:"organization_id"`
	ProjectSlug            string         `db:"project_slug"`
	IntegrationID          sql.NullString `db:"integration_id"`
	InstallationID         sql.NullString `db:"installation_id"`
	AccountName            sql.NullString `db:"account_name"`
	AccountType            sql.NullString `db:"account_type"`
	ExistingProjectID      sql.NullString `db:"existing_project_id"`
	ExistingIntegrationID  sql.NullString `db:"existing_integration_id"`
	ExistingRepositoryID   sql.NullString `db:"existing_repository_id"`
	ExistingRepositoryName sql.NullString `db:"existing_repository_name"`
	ExistingFullName       sql.NullString `db:"existing_full_name"`
	ExistingDefaultBranch  sql.NullString `db:"existing_default_branch"`
	ExistingRepositoryURL  sql.NullString `db:"existing_repository_url"`
	ExistingPrivate        sql.NullBool   `db:"existing_private"`
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateState(ctx context.Context, stateHash, organizationID, userID string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		WITH expired_states AS (
			DELETE FROM tbl_github_installation_state
			WHERE expires_at <= NOW()
			RETURNING id
		)
		INSERT INTO tbl_github_installation_state (state_hash, organization_id, user_id, expires_at)
		VALUES ($1, $2, $3, $4)
	`, stateHash, organizationID, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("create GitHub installation state: %w", err)
	}
	return nil
}

func (r *repository) FindState(ctx context.Context, stateHash string) (State, error) {
	var state State
	err := r.db.GetContext(ctx, &state, `
		SELECT organization_id, user_id
		FROM tbl_github_installation_state
		WHERE state_hash = $1 AND expires_at > NOW()
	`, stateHash)
	if errors.Is(err, sql.ErrNoRows) {
		return State{}, ErrStateInvalid
	}
	if err != nil {
		return State{}, fmt.Errorf("find GitHub installation state: %w", err)
	}
	return state, nil
}

func (r *repository) CompleteInstallation(ctx context.Context, stateHash string, installation Installation) (Integration, error) {
	var result Integration
	err := appdatabase.WithTx(ctx, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		state, err := lockInstallationState(ctx, tx, stateHash)
		if err != nil {
			return err
		}
		if state.consumedAt.Valid {
			result, err = findCompletedIntegration(ctx, tx, state.previousIntegrationID.String)
			if err != nil {
				return err
			}
			return nil
		}

		result, err = saveInstallation(ctx, tx, state.organizationID, state.userID, installation)
		if err != nil {
			return err
		}
		return consumeInstallationState(ctx, tx, stateHash, result.ID)
	})
	if err != nil {
		return Integration{}, err
	}
	return result, nil
}

func lockInstallationState(ctx context.Context, tx *sqlx.Tx, stateHash string) (installationState, error) {
	var state installationState
	err := tx.QueryRowxContext(ctx, `
		SELECT organization_id, user_id, consumed_at, integration_id
		FROM tbl_github_installation_state
		WHERE state_hash = $1 AND expires_at > NOW()
		FOR UPDATE
	`, stateHash).Scan(&state.organizationID, &state.userID, &state.consumedAt, &state.previousIntegrationID)
	if errors.Is(err, sql.ErrNoRows) {
		return installationState{}, ErrStateInvalid
	}
	if err != nil {
		return installationState{}, fmt.Errorf("lock GitHub installation state: %w", err)
	}
	return state, nil
}

func findCompletedIntegration(ctx context.Context, tx *sqlx.Tx, integrationID string) (Integration, error) {
	var result Integration
	err := tx.GetContext(ctx, &result, integrationSelect+`
		WHERE i.id = $1 AND p.code = 'GITHUB'
	`, integrationID)
	if errors.Is(err, sql.ErrNoRows) {
		return Integration{}, ErrStateInvalid
	}
	if err != nil {
		return Integration{}, fmt.Errorf("load previously completed GitHub integration: %w", err)
	}
	return result, nil
}

func saveInstallation(ctx context.Context, tx *sqlx.Tx, organizationID, userID string, installation Installation) (Integration, error) {
	var result Integration
	err := tx.GetContext(ctx, &result, `
		INSERT INTO tbl_organization_integration (
			organization_id, provider_id, external_account_id, external_account_name,
			external_account_type, installation_id, status, connected_by,
			connected_at, disconnected_at, provider_metadata
		)
		SELECT $1, p.id, $2, $3, $4, $5, 'ACTIVE', $6, NOW(), NULL, '{}'::jsonb
		FROM tbl_integration_provider p
		WHERE p.code = 'GITHUB' AND p.is_active = TRUE
		ON CONFLICT (organization_id, provider_id, external_account_id)
		DO UPDATE SET
			external_account_name = EXCLUDED.external_account_name,
			external_account_type = EXCLUDED.external_account_type,
			installation_id = EXCLUDED.installation_id,
			status = 'ACTIVE',
			connected_by = EXCLUDED.connected_by,
			connected_at = NOW(),
			disconnected_at = NULL,
			updated_at = NOW(),
			provider_metadata = '{}'::jsonb
		RETURNING id, organization_id, external_account_id, external_account_name,
			external_account_type, installation_id, status, connected_at
	`, organizationID, fmt.Sprint(installation.AccountID), installation.AccountName,
		installation.AccountType, fmt.Sprint(installation.ID), userID)
	if err != nil {
		return Integration{}, fmt.Errorf("save organization GitHub installation: %w", err)
	}
	result.ProviderCode = "GITHUB"
	return result, nil
}

func consumeInstallationState(ctx context.Context, tx *sqlx.Tx, stateHash, integrationID string) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE tbl_github_installation_state
		SET consumed_at = NOW(), integration_id = $2
		WHERE state_hash = $1 AND consumed_at IS NULL
	`, stateHash, integrationID); err != nil {
		return fmt.Errorf("consume GitHub installation state: %w", err)
	}
	return nil
}

const integrationSelect = `
	SELECT i.id, i.organization_id, p.code AS provider_code, i.external_account_id,
		i.external_account_name, COALESCE(i.external_account_type, '') AS external_account_type,
		COALESCE(i.installation_id, '') AS installation_id, i.status, i.connected_at
	FROM tbl_organization_integration i
	INNER JOIN tbl_integration_provider p ON p.id = i.provider_id
`

func (r *repository) ListGitHub(ctx context.Context, organizationID string) ([]Integration, error) {
	items := make([]Integration, 0)
	err := r.db.SelectContext(ctx, &items, integrationSelect+`
		WHERE i.organization_id = $1 AND p.code = 'GITHUB'
		ORDER BY i.connected_at DESC
	`, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list organization GitHub integrations: %w", err)
	}
	return items, nil
}

func (r *repository) FindExportContext(ctx context.Context, organizationID, projectID string) (ExportContext, error) {
	var row exportContextRow
	err := r.db.GetContext(ctx, &row, `
		SELECT p.id AS project_id, p.organization_id, p.slug AS project_slug,
			i.id AS integration_id, i.installation_id,
			i.external_account_name AS account_name, i.external_account_type AS account_type,
			pr.project_id AS existing_project_id, pr.integration_id AS existing_integration_id,
			pr.external_repository_id AS existing_repository_id,
			pr.repository_name AS existing_repository_name, pr.full_name AS existing_full_name,
			pr.default_branch AS existing_default_branch, pr.repository_url AS existing_repository_url,
			pr.is_private AS existing_private
		FROM tbl_project p
		LEFT JOIN LATERAL (
			SELECT oi.id, oi.installation_id, oi.external_account_name, oi.external_account_type
			FROM tbl_organization_integration oi
			INNER JOIN tbl_integration_provider provider ON provider.id = oi.provider_id
			WHERE oi.organization_id = p.organization_id AND oi.status = 'ACTIVE'
				AND provider.code = 'GITHUB'
			ORDER BY oi.connected_at DESC
			LIMIT 1
		) i ON TRUE
		LEFT JOIN tbl_project_repository pr
			ON pr.project_id = p.id AND pr.organization_id = p.organization_id
		WHERE p.id = $1 AND p.organization_id = $2 AND p.deleted_at IS NULL
	`, projectID, organizationID)
	if errors.Is(err, sql.ErrNoRows) {
		return ExportContext{}, ErrProjectNotFound
	}
	if err != nil {
		return ExportContext{}, fmt.Errorf("find project GitHub export context: %w", err)
	}
	if !row.IntegrationID.Valid {
		return ExportContext{}, ErrIntegrationNotFound
	}

	result := ExportContext{
		ProjectID: row.ProjectID, OrganizationID: row.OrganizationID, ProjectSlug: row.ProjectSlug,
		IntegrationID: row.IntegrationID.String, InstallationID: row.InstallationID.String,
		AccountName: row.AccountName.String, AccountType: row.AccountType.String,
	}
	if row.ExistingProjectID.Valid {
		result.Existing = &ExportRepository{
			ProjectID: row.ExistingProjectID.String, IntegrationID: row.ExistingIntegrationID.String,
			RepositoryID: row.ExistingRepositoryID.String, RepositoryName: row.ExistingRepositoryName.String,
			FullName: row.ExistingFullName.String, DefaultBranch: row.ExistingDefaultBranch.String,
			RepositoryURL: row.ExistingRepositoryURL.String, Private: row.ExistingPrivate.Bool,
		}
	}
	return result, nil
}

func (r *repository) SaveExportRepository(ctx context.Context, projectID, integrationID string, repository ExportRepository, userID string) (ExportRepository, error) {
	var result ExportRepository
	err := r.db.GetContext(ctx, &result, `
		INSERT INTO tbl_project_repository (
			organization_id, project_id, integration_id, external_repository_id,
			repository_name, full_name, default_branch, repository_url,
			is_private, linked_by
		)
		SELECT p.organization_id, p.id, $2, $3, $4, $5, $6, $7, $8, $9
		FROM tbl_project p
		WHERE p.id = $1 AND p.deleted_at IS NULL
		RETURNING project_id, integration_id, external_repository_id, repository_name,
			full_name, COALESCE(default_branch, '') AS default_branch,
			COALESCE(repository_url, '') AS repository_url, is_private
	`, projectID, integrationID, repository.RepositoryID, repository.RepositoryName,
		repository.FullName, repository.DefaultBranch, repository.RepositoryURL, repository.Private, userID)
	if err != nil {
		return ExportRepository{}, fmt.Errorf("save project GitHub export repository: %w", err)
	}
	return result, nil
}

func (r *repository) FindExportRepository(ctx context.Context, organizationID, projectID string) (ExportRepository, error) {
	var result ExportRepository
	err := r.db.GetContext(ctx, &result, `
		SELECT project_id, integration_id, external_repository_id, repository_name,
			full_name, COALESCE(default_branch, '') AS default_branch,
			COALESCE(repository_url, '') AS repository_url, is_private
		FROM tbl_project_repository
		WHERE organization_id = $1 AND project_id = $2
	`, organizationID, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return ExportRepository{}, ErrExportRepositoryNotFound
	}
	if err != nil {
		return ExportRepository{}, fmt.Errorf("find project GitHub export repository: %w", err)
	}
	return result, nil
}

func (r *repository) FindExportCredentials(ctx context.Context, organizationID, projectID string) (string, string, error) {
	var result struct {
		InstallationID string `db:"installation_id"`
		RepositoryURL  string `db:"repository_url"`
	}
	err := r.db.GetContext(ctx, &result, `
		SELECT i.installation_id, pr.repository_url
		FROM tbl_project_repository pr
		INNER JOIN tbl_organization_integration i
			ON i.id = pr.integration_id AND i.organization_id = pr.organization_id
		INNER JOIN tbl_integration_provider provider ON provider.id = i.provider_id
		WHERE pr.organization_id = $1 AND pr.project_id = $2
			AND i.status = 'ACTIVE' AND provider.code = 'GITHUB'
	`, organizationID, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrExportRepositoryNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("find project GitHub export credentials: %w", err)
	}
	return result.InstallationID, result.RepositoryURL, nil
}

func (r *repository) DisconnectGitHub(ctx context.Context, organizationID, integrationID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tbl_organization_integration i
		SET status = 'REVOKED', disconnected_at = COALESCE(disconnected_at, NOW()), updated_at = NOW()
		FROM tbl_integration_provider p
		WHERE i.provider_id = p.id AND p.code = 'GITHUB'
		  AND i.organization_id = $1 AND i.id = $2
	`, organizationID, integrationID)
	if err != nil {
		return fmt.Errorf("disconnect organization GitHub integration: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check GitHub disconnection result: %w", err)
	}
	if rows == 0 {
		return ErrIntegrationNotFound
	}
	return nil
}

func (r *repository) SetInstallationStatus(ctx context.Context, installationID, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tbl_organization_integration
		SET status = $1,
			disconnected_at = CASE WHEN $1 = 'ACTIVE' THEN NULL ELSE NOW() END,
			updated_at = NOW()
		WHERE installation_id = $2 AND status <> 'REVOKED'
	`, status, installationID)
	if err != nil {
		return fmt.Errorf("synchronize GitHub installation status: %w", err)
	}
	return nil
}
